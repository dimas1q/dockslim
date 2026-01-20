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

	recommendations := BuildRecommendations(layers, 0, "")

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

	recommendations := BuildRecommendations(layers, 0, "")

	if !hasRecommendation(recommendations, "too-many-layers") {
		t.Fatalf("expected too-many-layers recommendation")
	}
}

func TestBuildRecommendationsLargeImage(t *testing.T) {
	recommendations := BuildRecommendations(nil, largeImageThresholdBytes+1, "")

	if !hasRecommendation(recommendations, "large-image") {
		t.Fatalf("expected large-image recommendation")
	}
}

func TestBuildRecommendationsHugeTotalSize(t *testing.T) {
	recommendations := BuildRecommendations(nil, hugeImageThresholdBytes+1, "")

	if !hasRecommendation(recommendations, "huge-total-size") {
		t.Fatalf("expected huge-total-size recommendation")
	}
}

func TestBuildRecommendationsManyMediumLayers(t *testing.T) {
	layers := make([]registry.ManifestLayer, manyMediumLayersThreshold+1)
	for i := range layers {
		layers[i] = registry.ManifestLayer{
			Digest: "sha256:layer",
			Size:   manyMediumLayerMinBytes,
		}
	}

	recommendations := BuildRecommendations(layers, 0, "")

	if !hasRecommendation(recommendations, "many-medium-layers") {
		t.Fatalf("expected many-medium-layers recommendation")
	}
}

func TestBuildRecommendationsGzipLayersDetected(t *testing.T) {
	layers := []registry.ManifestLayer{
		{Digest: "sha256:layer", Size: 1024, MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip"},
	}

	recommendations := BuildRecommendations(layers, 0, "")

	if !hasRecommendation(recommendations, "gzip-layers-detected") {
		t.Fatalf("expected gzip-layers-detected recommendation")
	}
}

func TestBuildRecommendationsOciManifest(t *testing.T) {
	recommendations := BuildRecommendations(nil, 0, "application/vnd.oci.image.manifest.v1+json")

	if !hasRecommendation(recommendations, "oci-manifest") {
		t.Fatalf("expected oci-manifest recommendation")
	}
}

func TestBuildRecommendationsPullTimeRisk(t *testing.T) {
	layers := []registry.ManifestLayer{
		{Digest: "sha256:layer", Size: pullTimeRiskLayerThresholdBytes + 1},
	}

	recommendations := BuildRecommendations(layers, 0, "")

	if !hasRecommendation(recommendations, "pull-time-risk") {
		t.Fatalf("expected pull-time-risk recommendation")
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
