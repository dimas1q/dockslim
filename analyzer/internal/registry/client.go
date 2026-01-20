package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}

type ManifestSummary struct {
	MediaType  string
	Layers     []ManifestLayer
	LayerCount int
	TotalSize  int64
}

type ManifestLayer struct {
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	MediaType string `json:"mediaType"`
}

var ErrUnsupportedManifest = errors.New("unsupported manifest media type")

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Ping(ctx context.Context, registryURL, username, password string) error {
	endpoint, err := buildRegistryURL(registryURL, "/v2/")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("registry ping failed with status %d", resp.StatusCode)
}

func (c *Client) FetchManifest(ctx context.Context, registryURL, image, tag, username, password string) (ManifestSummary, error) {
	manifestPath := path.Join("/v2", image, "manifests", tag)
	endpoint, err := buildRegistryURL(registryURL, manifestPath)
	if err != nil {
		return ManifestSummary{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ManifestSummary{}, err
	}
	req.Header.Set("Accept", strings.Join([]string{
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.image.manifest.v1+json",
	}, ", "))
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ManifestSummary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ManifestSummary{}, HTTPStatusError{StatusCode: resp.StatusCode}
	}

	return parseManifest(resp.Body, resp.Header.Get("Content-Type"))
}

type HTTPStatusError struct {
	StatusCode int
}

func (err HTTPStatusError) Error() string {
	return fmt.Sprintf("manifest fetch failed with status %d", err.StatusCode)
}

func parseManifest(body io.Reader, contentType string) (ManifestSummary, error) {
	var manifest struct {
		SchemaVersion int             `json:"schemaVersion"`
		MediaType     string          `json:"mediaType"`
		Config        json.RawMessage `json:"config"`
		Layers        []ManifestLayer `json:"layers"`
	}

	if err := json.NewDecoder(body).Decode(&manifest); err != nil {
		return ManifestSummary{}, err
	}

	mediaType := normalizeMediaType(contentType)
	if mediaType == "" {
		mediaType = manifest.MediaType
	}

	if !isSupportedManifest(mediaType) {
		return ManifestSummary{}, ErrUnsupportedManifest
	}

	var total int64
	for _, layer := range manifest.Layers {
		total += layer.Size
	}

	return ManifestSummary{
		MediaType:  mediaType,
		Layers:     manifest.Layers,
		LayerCount: len(manifest.Layers),
		TotalSize:  total,
	}, nil
}

func normalizeMediaType(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ";")
	return strings.TrimSpace(parts[0])
}

func isSupportedManifest(mediaType string) bool {
	switch mediaType {
	case "application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.image.manifest.v1+json":
		return true
	default:
		return false
	}
}

func buildRegistryURL(base, endpoint string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	parsed.Path = joinRegistryPath(parsed.Path, endpoint)
	return parsed.String(), nil
}

func joinRegistryPath(basePath, endpoint string) string {
	basePath = strings.TrimSuffix(basePath, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")

	switch {
	case basePath == "" && endpoint == "":
		return "/"
	case basePath == "":
		basePath = "/"
	}

	if basePath == "/" {
		basePath = ""
	}

	combined := basePath
	if endpoint != "" {
		combined = basePath + "/" + endpoint
	}

	if strings.HasSuffix(endpoint, "/") && !strings.HasSuffix(combined, "/") {
		combined += "/"
	}

	if combined == "" {
		return "/"
	}

	if !strings.HasPrefix(combined, "/") {
		return "/" + combined
	}

	return combined
}
