package analyses

import (
	"net/url"
	"strings"
)

func normalizeImageReference(image, registryURL string) (string, error) {
	if strings.Contains(image, "://") {
		return "", ErrInvalidImage
	}

	parts := strings.SplitN(image, "/", 2)
	if len(parts) < 2 {
		return image, nil
	}

	hostPart := parts[0]
	if !looksLikeRegistryHost(hostPart) {
		return image, nil
	}

	registryHost, err := extractRegistryHostname(registryURL)
	if err != nil {
		return "", ErrInvalidRegistry
	}

	imageHost := extractImageHostname(hostPart)
	if !strings.EqualFold(imageHost, registryHost) {
		return "", ErrRegistryMismatch
	}

	if strings.TrimSpace(parts[1]) == "" {
		return "", ErrInvalidImage
	}

	return parts[1], nil
}

func looksLikeRegistryHost(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, ".") || strings.Contains(lower, ":") || lower == "localhost"
}

func extractRegistryHostname(registryURL string) (string, error) {
	parsed, err := url.Parse(registryURL)
	if err != nil {
		return "", err
	}
	if parsed.Hostname() == "" {
		return "", ErrInvalidRegistry
	}
	return parsed.Hostname(), nil
}

func extractImageHostname(hostPart string) string {
	if strings.HasPrefix(hostPart, "[") {
		return strings.Trim(hostPart, "[]")
	}
	if strings.Contains(hostPart, ":") {
		parts := strings.Split(hostPart, ":")
		return parts[0]
	}
	return hostPart
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
