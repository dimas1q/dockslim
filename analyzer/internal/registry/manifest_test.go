package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseManifestDockerSchema2(t *testing.T) {
	manifestJSON := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 7023,
			"digest": "sha256:config"
		},
		"layers": [
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 100,
				"digest": "sha256:layer1"
			},
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 200,
				"digest": "sha256:layer2"
			}
		]
	}`

	summary, err := parseManifest(strings.NewReader(manifestJSON), "application/vnd.docker.distribution.manifest.v2+json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary.LayerCount != 2 {
		t.Fatalf("expected 2 layers, got %d", summary.LayerCount)
	}
	if summary.TotalSize != 300 {
		t.Fatalf("expected total size 300, got %d", summary.TotalSize)
	}
	if summary.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		t.Fatalf("unexpected media type %s", summary.MediaType)
	}
	if summary.Layers[0].Digest != "sha256:layer1" {
		t.Fatalf("unexpected digest %s", summary.Layers[0].Digest)
	}
}

func TestParseManifestListDocker(t *testing.T) {
	listJSON := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
		"manifests": [
			{
				"digest": "sha256:linux-amd64",
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"platform": {"architecture": "amd64", "os": "linux"}
			},
			{
				"digest": "sha256:linux-arm64",
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"platform": {"architecture": "arm64", "os": "linux"}
			}
		]
	}`

	list, err := parseManifestList([]byte(listJSON))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if list.MediaType != "application/vnd.docker.distribution.manifest.list.v2+json" {
		t.Fatalf("unexpected media type %s", list.MediaType)
	}
	if len(list.Manifests) != 2 {
		t.Fatalf("expected 2 manifests, got %d", len(list.Manifests))
	}
}

func TestSelectManifestDigestPrefersLinuxAmd64(t *testing.T) {
	manifests := []ManifestDescriptor{
		{
			Digest: "sha256:arm64",
			Platform: struct {
				OS           string `json:"os"`
				Architecture string `json:"architecture"`
				Variant      string `json:"variant,omitempty"`
			}{
				OS:           "linux",
				Architecture: "arm64",
			},
		},
		{
			Digest: "sha256:amd64",
			Platform: struct {
				OS           string `json:"os"`
				Architecture string `json:"architecture"`
				Variant      string `json:"variant,omitempty"`
			}{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
	}

	selected := selectManifestDigest(manifests)
	if selected != "sha256:amd64" {
		t.Fatalf("expected amd64 digest, got %s", selected)
	}
}

func TestFetchManifestFromList(t *testing.T) {
	listJSON := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
		"manifests": [
			{
				"digest": "sha256:linux-amd64",
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"platform": {"architecture": "amd64", "os": "linux"}
			},
			{
				"digest": "sha256:linux-arm64",
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"platform": {"architecture": "arm64", "os": "linux"}
			}
		]
	}`

	manifestJSON := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"layers": [
			{"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip", "size": 10, "digest": "sha256:layer1"},
			{"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip", "size": 20, "digest": "sha256:layer2"}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/example/app/manifests/latest":
			w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.list.v2+json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(listJSON))
		case "/v2/example/app/manifests/sha256:linux-amd64":
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "manifest.list.v2+json") {
				t.Fatalf("expected accept header without list types, got %s", accept)
			}
			w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(manifestJSON))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient()
	summary, err := client.FetchManifest(context.Background(), server.URL, "example/app", "latest", "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.LayerCount != 2 {
		t.Fatalf("expected 2 layers, got %d", summary.LayerCount)
	}
	if summary.TotalSize != 30 {
		t.Fatalf("expected total size 30, got %d", summary.TotalSize)
	}
}
