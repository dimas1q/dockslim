package analysis

import (
	"sort"

	"github.com/dimas1q/dockslim/analyzer/internal/registry"
)

const (
	smallLayerThresholdBytes = 1 * 1024 * 1024
	manySmallLayersThreshold = 20
)

type Recommendation struct {
	ID              string `json:"id"`
	Severity        string `json:"severity"`
	Category        string `json:"category"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	SuggestedAction string `json:"suggested_action"`
}

func BuildRecommendations(layers []registry.ManifestLayer, totalSize int64) []Recommendation {
	recommendations := make([]Recommendation, 0)

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

	if len(layers) > manyLayersThreshold {
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

	sort.SliceStable(recommendations, func(i, j int) bool {
		return severityRank(recommendations[i].Severity) < severityRank(recommendations[j].Severity)
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
