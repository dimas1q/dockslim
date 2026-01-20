package analysis

import (
	"testing"

	"github.com/dimas1q/dockslim/analyzer/internal/registry"
)

func TestBuildRecommendationsLargeLayer(t *testing.T) {
	layers := []registry.ManifestLayer{
		{Digest: "sha256:small", Size: 1024},
		{Digest: "sha256:big", Size: largeLayerThresholdBytes + 1},
	}

	recommendations := BuildRecommendations(layers, 0)

	if !hasRecommendation(recommendations, "large-layer") {
		t.Fatalf("expected large-layer recommendation")
	}
}

func TestBuildRecommendationsTooManyLayers(t *testing.T) {
	layers := make([]registry.ManifestLayer, manyLayersThreshold+1)
	for i := range layers {
		layers[i] = registry.ManifestLayer{
			Digest: "sha256:layer",
			Size:   1024,
		}
	}

	recommendations := BuildRecommendations(layers, 0)

	if !hasRecommendation(recommendations, "too-many-layers") {
		t.Fatalf("expected too-many-layers recommendation")
	}
}

func TestBuildRecommendationsLargeImage(t *testing.T) {
	recommendations := BuildRecommendations(nil, largeImageThresholdBytes+1)

	if !hasRecommendation(recommendations, "large-image") {
		t.Fatalf("expected large-image recommendation")
	}
}

func hasRecommendation(recommendations []Recommendation, id string) bool {
	for _, recommendation := range recommendations {
		if recommendation.ID == id {
			return true
		}
	}
	return false
}
