package analysis

import (
	"sort"
	"strings"

	"github.com/dimas1q/dockslim/analyzer/internal/registry"
)

type Recommendation struct {
	ID              string `json:"id"`
	Severity        string `json:"severity"`
	Category        string `json:"category"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	SuggestedAction string `json:"suggested_action"`
}

func BuildRecommendations(layers []registry.ManifestLayer, totalSize int64, manifestMediaType string) []Recommendation {
	recommendations := make([]Recommendation, 0)
	layerCount := len(layers)
	largestLayerSize := int64(0)
	mediumLayerCount := 0
	hasGzipLayer := false

	for _, layer := range layers {
		if layer.Size > largestLayerSize {
			largestLayerSize = layer.Size
		}
		if layer.Size >= manyMediumLayerMinBytes && layer.Size <= manyMediumLayerMaxBytes {
			mediumLayerCount++
		}
		if strings.Contains(layer.MediaType, "gzip") {
			hasGzipLayer = true
		}
	}

	for _, layer := range layers {
		if layer.Size > largeLayerThresholdBytes {
			recommendations = append(recommendations, Recommendation{
				ID:              "large-layer",
				Severity:        "critical",
				Category:        "layers",
				Title:           "Very large image layer detected",
				Description:     "One or more layers exceed 200 MB, significantly increasing image size.",
				SuggestedAction: "Split large RUN steps and clean package caches in the same layer.",
			})
			break
		}
	}

	if layerCount > manyLayersThreshold {
		recommendations = append(recommendations, Recommendation{
			ID:              "too-many-layers",
			Severity:        "warning",
			Category:        "layers",
			Title:           "High number of image layers",
			Description:     "The image contains many layers, which can slow pulls and cache efficiency.",
			SuggestedAction: "Combine RUN instructions where possible.",
		})
	}

	if totalSize > largeImageThresholdBytes {
		recommendations = append(recommendations, Recommendation{
			ID:              "large-image",
			Severity:        "warning",
			Category:        "size",
			Title:           "Large container image size",
			Description:     "The total image size exceeds 1 GB.",
			SuggestedAction: "Consider using a slimmer base image (alpine, distroless).",
		})
	}

	smallLayers := 0
	for _, layer := range layers {
		if layer.Size < smallLayerThresholdBytes {
			smallLayers++
		}
	}
	if smallLayers > manySmallLayersThreshold {
		recommendations = append(recommendations, Recommendation{
			ID:              "many-small-layers",
			Severity:        "info",
			Category:        "layers",
			Title:           "Many very small layers",
			Description:     "The image contains many very small layers.",
			SuggestedAction: "Reduce COPY granularity to improve layer efficiency.",
		})
	}

	if totalSize > hugeImageThresholdBytes {
		recommendations = append(recommendations, Recommendation{
			ID:              "huge-total-size",
			Severity:        "critical",
			Category:        "size",
			Title:           "Extremely large image",
			Description:     "The total image size exceeds 2 GB, which will slow pulls and deployments.",
			SuggestedAction: "Use distroless or slim bases, remove build dependencies, and apply multi-stage builds.",
		})
	}

	if mediumLayerCount > manyMediumLayersThreshold {
		recommendations = append(recommendations, Recommendation{
			ID:              "many-medium-layers",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Many medium-sized layers",
			Description:     "Dozens of 10–50 MB layers can slow downloads and reduce cache efficiency.",
			SuggestedAction: "Consolidate install steps and limit intermediate layers.",
		})
	}

	if layerCount > 0 && layerCount <= tooFewLayersThreshold && totalSize > hugeFewLayersImageThresholdBytes {
		recommendations = append(recommendations, Recommendation{
			ID:              "too-few-layers-but-huge",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Few layers but very large image",
			Description:     "A small number of layers holds most of the image size, which often means a giant RUN step.",
			SuggestedAction: "Split steps and clean caches in the same layer to keep size down.",
		})
	}

	if hasGzipLayer {
		recommendations = append(recommendations, Recommendation{
			ID:              "gzip-layers-detected",
			Severity:        "info",
			Category:        "layers",
			Title:           "Gzip-compressed layers detected",
			Description:     "The image uses gzip-compressed layers.",
			SuggestedAction: "Consider modern compression settings (BuildKit) for better pull performance.",
		})
	}

	switch manifestMediaType {
	case "application/vnd.oci.image.manifest.v1+json":
		recommendations = append(recommendations, Recommendation{
			ID:              "oci-manifest",
			Severity:        "info",
			Category:        "base-image",
			Title:           "OCI manifest detected",
			Description:     "OCI manifests are widely supported but still worth validating in your tooling.",
			SuggestedAction: "Ensure your CI/CD and runtime tooling fully support OCI images.",
		})
	case "application/vnd.docker.distribution.manifest.v2+json":
		recommendations = append(recommendations, Recommendation{
			ID:              "docker-manifest",
			Severity:        "info",
			Category:        "base-image",
			Title:           "Docker schema2 manifest detected",
			Description:     "Docker schema2 is common, but OCI offers broader portability.",
			SuggestedAction: "Consider migrating to OCI manifests where supported.",
		})
	}

	if largestLayerSize > vendoredLayerThresholdBytes && layerCount < vendoredLayerMaxCountThreshold {
		recommendations = append(recommendations, Recommendation{
			ID:              "likely-vendor-in-image",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Large dependency bundle likely baked in",
			Description:     "A single layer exceeds 300 MB while the layer count is low.",
			SuggestedAction: "Avoid vendoring large dependencies into the runtime image; use a builder stage.",
		})
	}

	if largestLayerSize > cacheCleanupLayerThresholdBytes {
		recommendations = append(recommendations, Recommendation{
			ID:              "cache-cleanup",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Large layer suggests package cache bloat",
			Description:     "At least one layer exceeds 150 MB, which often means package caches were left behind.",
			SuggestedAction: "Clean apt/yum/apk caches in the same RUN instruction as installs.",
		})
	}

	if totalSize > rebuildFrequentlyThresholdBytes {
		recommendations = append(recommendations, Recommendation{
			ID:              "rebuild-frequently",
			Severity:        "info",
			Category:        "size",
			Title:           "Consider frequent rebuilds with caching",
			Description:     "Images over 500 MB benefit from aggressive layer caching and pinned base digests.",
			SuggestedAction: "Enable CI layer caching and pin base images by digest.",
		})
	}

	if totalSize > pullTimeRiskTotalThresholdBytes || largestLayerSize > pullTimeRiskLayerThresholdBytes {
		recommendations = append(recommendations, Recommendation{
			ID:              "pull-time-risk",
			Severity:        "warning",
			Category:        "size",
			Title:           "High pull-time risk",
			Description:     "Very large images or layers can cause slow pulls and timeouts.",
			SuggestedAction: "Keep registries close to workloads, use slim bases, and reduce oversized layers.",
		})
	}

	sort.SliceStable(recommendations, func(i, j int) bool {
		left := severityRank(recommendations[i].Severity)
		right := severityRank(recommendations[j].Severity)
		if left != right {
			return left < right
		}
		return recommendations[i].ID < recommendations[j].ID
	})

	return recommendations
}

func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 3
	}
}
