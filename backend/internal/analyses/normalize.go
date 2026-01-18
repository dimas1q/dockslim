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

	registryHost, err := extractRegistryHost(registryURL)
	if err != nil {
		return "", ErrInvalidRegistry
	}

	if !strings.EqualFold(hostPart, registryHost) {
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

func extractRegistryHost(registryURL string) (string, error) {
	parsed, err := url.Parse(registryURL)
	if err != nil {
		return "", err
	}
	if parsed.Host == "" {
		return "", ErrInvalidRegistry
	}
	return parsed.Host, nil
}
