package registry

import (
	"context"
	"encoding/json"
	"fmt"
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
	LayerCount int
	TotalSize  int64
}

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
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
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

	var manifest struct {
		SchemaVersion int `json:"schemaVersion"`
		Config        struct {
			Size int64 `json:"size"`
		} `json:"config"`
		Layers []struct {
			Size int64 `json:"size"`
		} `json:"layers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return ManifestSummary{}, err
	}

	var total int64
	total += manifest.Config.Size
	for _, layer := range manifest.Layers {
		total += layer.Size
	}

	return ManifestSummary{
		LayerCount: len(manifest.Layers),
		TotalSize:  total,
	}, nil
}

type HTTPStatusError struct {
	StatusCode int
}

func (err HTTPStatusError) Error() string {
	return fmt.Sprintf("manifest fetch failed with status %d", err.StatusCode)
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
