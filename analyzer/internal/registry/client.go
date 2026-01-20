package registry

import (
	"bytes"
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

type ManifestDescriptor struct {
	Digest    string `json:"digest"`
	MediaType string `json:"mediaType"`
	Platform  struct {
		OS           string `json:"os"`
		Architecture string `json:"architecture"`
		Variant      string `json:"variant,omitempty"`
	} `json:"platform"`
}

type ManifestList struct {
	MediaType string               `json:"mediaType"`
	Manifests []ManifestDescriptor `json:"manifests"`
}

var ErrUnsupportedManifest = errors.New("unsupported manifest media type")
var ErrNoSupportedManifest = errors.New("no supported manifest found in list")

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
	return c.fetchManifest(ctx, registryURL, image, tag, username, password, true)
}

func (c *Client) fetchManifest(ctx context.Context, registryURL, image, ref, username, password string, includeLists bool) (ManifestSummary, error) {
	manifestPath := path.Join("/v2", image, "manifests", ref)
	endpoint, err := buildRegistryURL(registryURL, manifestPath)
	if err != nil {
		return ManifestSummary{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ManifestSummary{}, err
	}
	req.Header.Set("Accept", strings.Join(buildManifestAcceptHeader(includeLists), ", "))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ManifestSummary{}, err
	}

	mediaType := resolveMediaType(body, resp.Header.Get("Content-Type"))
	if isManifestList(mediaType) {
		manifestList, err := parseManifestList(body)
		if err != nil {
			return ManifestSummary{}, err
		}
		digest := selectManifestDigest(manifestList.Manifests)
		if digest == "" {
			return ManifestSummary{}, ErrNoSupportedManifest
		}
		return c.fetchManifest(ctx, registryURL, image, digest, username, password, false)
	}

	return parseManifestBytes(body, mediaType)
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

	mediaType := resolveMediaTypeFromManifest(contentType, manifest.MediaType)
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

func parseManifestBytes(body []byte, contentType string) (ManifestSummary, error) {
	return parseManifest(bytes.NewReader(body), contentType)
}

func parseManifestList(body []byte) (ManifestList, error) {
	var list ManifestList
	if err := json.Unmarshal(body, &list); err != nil {
		return ManifestList{}, err
	}
	if list.MediaType == "" {
		var wrapper struct {
			MediaType string `json:"mediaType"`
		}
		if err := json.Unmarshal(body, &wrapper); err == nil {
			list.MediaType = wrapper.MediaType
		}
	}
	if !isManifestList(list.MediaType) {
		return ManifestList{}, ErrUnsupportedManifest
	}
	return list, nil
}

func selectManifestDigest(manifests []ManifestDescriptor) string {
	supported := make([]ManifestDescriptor, 0, len(manifests))
	for _, manifest := range manifests {
		if isSupportedManifest(manifest.MediaType) {
			supported = append(supported, manifest)
		}
	}
	if len(supported) == 0 {
		return ""
	}
	for _, manifest := range manifests {
		if !isSupportedManifest(manifest.MediaType) {
			continue
		}
		if manifest.Platform.OS == "linux" && manifest.Platform.Architecture == "amd64" {
			return manifest.Digest
		}
	}
	for _, manifest := range manifests {
		if !isSupportedManifest(manifest.MediaType) {
			continue
		}
		if manifest.Platform.OS == "linux" && manifest.Platform.Architecture == "arm64" {
			return manifest.Digest
		}
	}
	return supported[0].Digest
}

func resolveMediaType(body []byte, contentType string) string {
	mediaType := normalizeMediaType(contentType)
	if mediaType != "" {
		return mediaType
	}
	var wrapper struct {
		MediaType string `json:"mediaType"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return ""
	}
	return wrapper.MediaType
}

func resolveMediaTypeFromManifest(contentType, manifestMediaType string) string {
	mediaType := normalizeMediaType(contentType)
	if mediaType != "" {
		return mediaType
	}
	return manifestMediaType
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

func isManifestList(mediaType string) bool {
	switch mediaType {
	case "application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.oci.image.index.v1+json":
		return true
	default:
		return false
	}
}

func buildManifestAcceptHeader(includeLists bool) []string {
	accept := []string{
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.image.manifest.v1+json",
	}
	if includeLists {
		accept = append(accept,
			"application/vnd.docker.distribution.manifest.list.v2+json",
			"application/vnd.oci.image.index.v1+json",
		)
	}
	return accept
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
