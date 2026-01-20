package registry

import (
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
