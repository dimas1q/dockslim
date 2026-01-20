package analysis

import (
	"sort"
	"strings"

	"github.com/dimas1q/dockslim/analyzer/internal/registry"
)

const (
	smallLayerThresholdBytes = 1 * 1024 * 1024
	manySmallLayersThreshold = 20
	mediumLayerMinBytes      = 10 * 1024 * 1024
	mediumLayerMaxBytes      = 50 * 1024 * 1024
	tinyLayerThresholdBytes  = 256 * 1024
	hugeImageThresholdBytes  = 2 * 1024 * 1024 * 1024
	pullRiskThresholdBytes   = 1536 * 1024 * 1024
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
	var (
		maxLayerSize       int64
		largeLayerCount    int
		mediumLayerCount   int
		smallLayerCount    int
		tinyLayerCount     int
		veryLargeLayerSeen bool
		hugeLayerSeen      bool
		pullRiskLayerSeen  bool
		gzipLayerSeen      bool
		uncompressedLayer  bool
	)

	for _, layer := range layers {
		if layer.Size > maxLayerSize {
			maxLayerSize = layer.Size
		}
		if layer.Size > largeLayerThresholdBytes {
			largeLayerCount++
		}
		if layer.Size >= mediumLayerMinBytes && layer.Size <= mediumLayerMaxBytes {
			mediumLayerCount++
		}
		if layer.Size < smallLayerThresholdBytes {
			smallLayerCount++
		}
		if layer.Size < tinyLayerThresholdBytes {
			tinyLayerCount++
		}
		if layer.Size > 150*1024*1024 {
			veryLargeLayerSeen = true
		}
		if layer.Size > 500*1024*1024 {
			hugeLayerSeen = true
		}
		if layer.Size > 400*1024*1024 {
			pullRiskLayerSeen = true
		}
		if strings.Contains(layer.MediaType, "gzip") {
			gzipLayerSeen = true
		}
		if strings.Contains(layer.MediaType, "tar") && !strings.Contains(layer.MediaType, "gzip") {
			uncompressedLayer = true
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

	if smallLayerCount > manySmallLayersThreshold {
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
			Title:           "Image size is extremely large",
			Description:     "The total image size exceeds 2 GB, which is costly to distribute and store.",
			SuggestedAction: "Adopt a slim or distroless base, remove build-only dependencies, and use multi-stage builds.",
		})
	}

	if mediumLayerCount > 25 {
		recommendations = append(recommendations, Recommendation{
			ID:              "many-medium-layers",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Many medium-sized layers",
			Description:     "The image has a high count of 10–50 MB layers, adding pull overhead.",
			SuggestedAction: "Consolidate install steps and reduce intermediate artifacts.",
		})
	}

	if layerCount <= 5 && totalSize > 800*1024*1024 {
		recommendations = append(recommendations, Recommendation{
			ID:              "too-few-layers-but-huge",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Few layers but very large image",
			Description:     "The image has very few layers but is still large, suggesting oversized RUN steps.",
			SuggestedAction: "Split large build steps and clean caches in the same layer.",
		})
	}

	if gzipLayerSeen {
		recommendations = append(recommendations, Recommendation{
			ID:              "gzip-layers-detected",
			Severity:        "info",
			Category:        "layers",
			Title:           "Compressed layers detected",
			Description:     "Layer compression is in use, which improves transfer time.",
			SuggestedAction: "Ensure your build tooling uses modern compression settings for consistent results.",
		})
	}

	if manifestMediaType == "application/vnd.oci.image.manifest.v1+json" {
		recommendations = append(recommendations, Recommendation{
			ID:              "oci-manifest",
			Severity:        "info",
			Category:        "base-image",
			Title:           "OCI manifest format",
			Description:     "The image uses the OCI manifest format.",
			SuggestedAction: "Confirm downstream tooling supports OCI to avoid compatibility surprises.",
		})
	}

	if manifestMediaType == "application/vnd.docker.distribution.manifest.v2+json" {
		recommendations = append(recommendations, Recommendation{
			ID:              "docker-manifest",
			Severity:        "info",
			Category:        "base-image",
			Title:           "Docker Schema 2 manifest format",
			Description:     "The image uses Docker schema 2 manifests.",
			SuggestedAction: "OCI migration is optional but can improve ecosystem compatibility.",
		})
	}

	if maxLayerSize > 300*1024*1024 && layerCount < 15 {
		recommendations = append(recommendations, Recommendation{
			ID:              "likely-vendor-in-image",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Large dependency bundle likely included",
			Description:     "A very large layer in a small layer set often indicates vendored dependencies in the runtime image.",
			SuggestedAction: "Move dependency compilation to a builder stage and copy only runtime artifacts.",
		})
	}

	if veryLargeLayerSeen {
		recommendations = append(recommendations, Recommendation{
			ID:              "cache-cleanup",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Large layers suggest cached artifacts",
			Description:     "One or more layers exceed 150 MB, which often includes package caches.",
			SuggestedAction: "Clean apt/yum/apk caches in the same RUN step that installs packages.",
		})
	}

	if totalSize > 500*1024*1024 {
		recommendations = append(recommendations, Recommendation{
			ID:              "rebuild-frequently",
			Severity:        "info",
			Category:        "size",
			Title:           "Large images benefit from cache hygiene",
			Description:     "Larger images are slower to rebuild and distribute.",
			SuggestedAction: "Enable CI layer caching and pin base image digests for repeatability.",
		})
	}

	if totalSize > pullRiskThresholdBytes || pullRiskLayerSeen {
		recommendations = append(recommendations, Recommendation{
			ID:              "pull-time-risk",
			Severity:        "warning",
			Category:        "size",
			Title:           "Pull time risk detected",
			Description:     "Very large images or layers increase timeouts and cold-start delays.",
			SuggestedAction: "Use regional registries, reduce layer sizes, and adopt slimmer bases.",
		})
	}

	if hugeLayerSeen {
		recommendations = append(recommendations, Recommendation{
			ID:              "single-mega-layer",
			Severity:        "critical",
			Category:        "layers",
			Title:           "Single extremely large layer detected",
			Description:     "A layer over 500 MB creates costly rebuilds and slow pulls.",
			SuggestedAction: "Split build steps and remove temporary artifacts before the layer is committed.",
		})
	}

	if totalSize > 200*1024*1024 && maxLayerSize > int64(float64(totalSize)*0.6) {
		recommendations = append(recommendations, Recommendation{
			ID:              "layer-size-skew",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Layer size is highly skewed",
			Description:     "A single layer accounts for the majority of the image size.",
			SuggestedAction: "Break the layer up and remove temporary build outputs.",
		})
	}

	if largeLayerCount >= 3 {
		recommendations = append(recommendations, Recommendation{
			ID:              "multiple-large-layers",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Multiple large layers detected",
			Description:     "Several layers are above 200 MB, increasing pull overhead.",
			SuggestedAction: "Consolidate steps and remove large caches to reduce repeated cost.",
		})
	}

	if tinyLayerCount > 50 {
		recommendations = append(recommendations, Recommendation{
			ID:              "too-many-tiny-layers",
			Severity:        "info",
			Category:        "layers",
			Title:           "Many tiny layers",
			Description:     "Numerous tiny layers add metadata overhead during pulls.",
			SuggestedAction: "Combine small COPY operations into fewer layers.",
		})
	}

	if totalSize > 700*1024*1024 && layerCount <= 12 {
		recommendations = append(recommendations, Recommendation{
			ID:              "base-image-heavy",
			Severity:        "warning",
			Category:        "base-image",
			Title:           "Base image likely oversized",
			Description:     "Large images with few layers often point to a heavyweight base image.",
			SuggestedAction: "Switch to a slimmer base image or distroless runtime.",
		})
	}

	if totalSize > 300*1024*1024 && layerCount <= 8 {
		recommendations = append(recommendations, Recommendation{
			ID:              "base-image-slim-option",
			Severity:        "info",
			Category:        "base-image",
			Title:           "Consider a slimmer base image",
			Description:     "The image is moderate in size with few layers.",
			SuggestedAction: "Evaluate alpine or slim variants of your base image.",
		})
	}

	if totalSize > 400*1024*1024 && layerCount > 12 {
		recommendations = append(recommendations, Recommendation{
			ID:              "base-image-refresh",
			Severity:        "info",
			Category:        "base-image",
			Title:           "Review base image freshness",
			Description:     "Large images with many layers often benefit from a refreshed base image.",
			SuggestedAction: "Update the base image digest and drop deprecated packages.",
		})
	}

	if layerCount > 80 {
		recommendations = append(recommendations, Recommendation{
			ID:              "excessive-layer-count",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Excessive number of layers",
			Description:     "Extremely high layer counts slow pulls and increase cache misses.",
			SuggestedAction: "Merge RUN steps and review build step granularity.",
		})
	}

	if layerCount > 25 && totalSize > 400*1024*1024 {
		recommendations = append(recommendations, Recommendation{
			ID:              "layer-churn-risk",
			Severity:        "info",
			Category:        "layers",
			Title:           "High layer count increases churn risk",
			Description:     "Many layers amplify cache invalidation when rebuilding.",
			SuggestedAction: "Group related file copies and installs into fewer layers.",
		})
	}

	if uncompressedLayer {
		recommendations = append(recommendations, Recommendation{
			ID:              "uncompressed-layers",
			Severity:        "warning",
			Category:        "layers",
			Title:           "Uncompressed layers detected",
			Description:     "Some layers appear to be uncompressed, inflating transfer size.",
			SuggestedAction: "Enable compression in your build tooling to reduce layer size.",
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
