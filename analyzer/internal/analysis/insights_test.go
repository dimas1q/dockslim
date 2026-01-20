package analysis

import (
	"testing"

	"github.com/dimas1q/dockslim/analyzer/internal/registry"
)

func TestBuildInsightsLargeLayerWarning(t *testing.T) {
	layers := []registry.ManifestLayer{
		{Digest: "sha256:small", Size: 1024},
		{Digest: "sha256:big", Size: largeLayerThresholdBytes + 1},
	}

	insights := BuildInsights(layers, layers[1].Size)

	if insights.LayerCount != 2 {
		t.Fatalf("expected 2 layers, got %d", insights.LayerCount)
	}
	if len(insights.LargestLayers) == 0 || insights.LargestLayers[0].Digest != "sha256:big" {
		t.Fatalf("expected largest layer to be sha256:big")
	}
	if len(insights.Warnings) == 0 {
		t.Fatalf("expected warnings, got none")
	}
}

func TestBuildInsightsManyLayersWarning(t *testing.T) {
	layers := make([]registry.ManifestLayer, manyLayersThreshold+1)
	for i := range layers {
		layers[i] = registry.ManifestLayer{
			Digest: "sha256:layer",
			Size:   1024,
		}
	}

	insights := BuildInsights(layers, 0)

	found := false
	for _, warning := range insights.Warnings {
		if warning == "Image contains many layers (>40)" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected many layers warning")
	}
}
