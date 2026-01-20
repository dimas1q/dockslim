package analysis

import (
	"sort"

	"github.com/dimas1q/dockslim/analyzer/internal/registry"
)

const maxLargestLayers = 3

type LayerInsight struct {
	Digest    string `json:"digest"`
	SizeBytes int64  `json:"size_bytes"`
}

type Insights struct {
	LayerCount    int            `json:"layer_count"`
	LargestLayers []LayerInsight `json:"largest_layers"`
	Warnings      []string       `json:"warnings"`
}

func BuildInsights(layers []registry.ManifestLayer, totalSize int64) Insights {
	insights := Insights{
		LayerCount: len(layers),
	}

	largest := make([]registry.ManifestLayer, len(layers))
	copy(largest, layers)
	sort.Slice(largest, func(i, j int) bool {
		return largest[i].Size > largest[j].Size
	})

	for i := 0; i < len(largest) && i < maxLargestLayers; i++ {
		insights.LargestLayers = append(insights.LargestLayers, LayerInsight{
			Digest:    largest[i].Digest,
			SizeBytes: largest[i].Size,
		})
	}

	for _, layer := range layers {
		if layer.Size > largeLayerThresholdBytes {
			insights.Warnings = append(insights.Warnings, "Image has very large layers (>200MB)")
			break
		}
	}

	if len(layers) > manyLayersThreshold {
		insights.Warnings = append(insights.Warnings, "Image contains many layers (>40)")
	}

	if totalSize > largeImageThresholdBytes {
		insights.Warnings = append(insights.Warnings, "Image is larger than 1GB")
	}

	return insights
}
