package analyses

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
)

type Comparison struct {
	ProjectID uuid.UUID           `json:"project_id"`
	Image     string              `json:"image"`
	From      ComparisonAnalysis  `json:"from"`
	To        ComparisonAnalysis  `json:"to"`
	Summary   ComparisonSummary   `json:"summary"`
	Layers    ComparisonLayerDiff `json:"layers"`
}

type ComparisonAnalysis struct {
	AnalysisID     uuid.UUID `json:"analysis_id"`
	Tag            string    `json:"tag"`
	CreatedAt      time.Time `json:"created_at"`
	TotalSizeBytes int64     `json:"total_size_bytes"`
	LayerCount     int       `json:"layer_count"`
}

type ComparisonSummary struct {
	TotalSizeDiffBytes int64 `json:"total_size_diff_bytes"`
	LayerCountDiff     int   `json:"layer_count_diff"`
}

type ComparisonLayerDiff struct {
	Added   []LayerDiff `json:"added"`
	Removed []LayerDiff `json:"removed"`
}

type LayerDiff struct {
	Digest    string `json:"digest"`
	SizeBytes int64  `json:"size_bytes"`
}

type analysisResult struct {
	Layers         []LayerDiff `json:"layers"`
	TotalSizeBytes *int64      `json:"total_size_bytes"`
}

func BuildComparison(from, to ImageAnalysis) (Comparison, error) {
	fromResult, err := parseAnalysisResult(from.ResultJSON)
	if err != nil {
		return Comparison{}, err
	}
	toResult, err := parseAnalysisResult(to.ResultJSON)
	if err != nil {
		return Comparison{}, err
	}

	fromLayers := normalizeLayers(fromResult.Layers)
	toLayers := normalizeLayers(toResult.Layers)

	added, removed := diffLayers(fromLayers, toLayers)

	fromTotal := resolveTotalSize(from, fromResult)
	toTotal := resolveTotalSize(to, toResult)

	comparison := Comparison{
		ProjectID: from.ProjectID,
		Image:     from.Image,
		From: ComparisonAnalysis{
			AnalysisID:     from.ID,
			Tag:            from.Tag,
			CreatedAt:      from.CreatedAt,
			TotalSizeBytes: fromTotal,
			LayerCount:     len(fromLayers),
		},
		To: ComparisonAnalysis{
			AnalysisID:     to.ID,
			Tag:            to.Tag,
			CreatedAt:      to.CreatedAt,
			TotalSizeBytes: toTotal,
			LayerCount:     len(toLayers),
		},
		Summary: ComparisonSummary{
			TotalSizeDiffBytes: toTotal - fromTotal,
			LayerCountDiff:     len(toLayers) - len(fromLayers),
		},
		Layers: ComparisonLayerDiff{
			Added:   added,
			Removed: removed,
		},
	}

	return comparison, nil
}

func parseAnalysisResult(raw json.RawMessage) (analysisResult, error) {
	if len(raw) == 0 {
		return analysisResult{}, nil
	}
	var result analysisResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return analysisResult{}, err
	}
	return result, nil
}

func normalizeLayers(layers []LayerDiff) []LayerDiff {
	if len(layers) == 0 {
		return nil
	}
	normalized := make([]LayerDiff, 0, len(layers))
	for _, layer := range layers {
		if layer.Digest == "" {
			continue
		}
		normalized = append(normalized, layer)
	}
	return normalized
}

func diffLayers(fromLayers, toLayers []LayerDiff) ([]LayerDiff, []LayerDiff) {
	fromMap := make(map[string]LayerDiff, len(fromLayers))
	for _, layer := range fromLayers {
		fromMap[layer.Digest] = layer
	}
	toMap := make(map[string]LayerDiff, len(toLayers))
	for _, layer := range toLayers {
		toMap[layer.Digest] = layer
	}

	added := make([]LayerDiff, 0)
	for digest, layer := range toMap {
		if _, ok := fromMap[digest]; !ok {
			added = append(added, layer)
		}
	}

	removed := make([]LayerDiff, 0)
	for digest, layer := range fromMap {
		if _, ok := toMap[digest]; !ok {
			removed = append(removed, layer)
		}
	}

	sortLayers := func(items []LayerDiff) {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].SizeBytes != items[j].SizeBytes {
				return items[i].SizeBytes > items[j].SizeBytes
			}
			return items[i].Digest < items[j].Digest
		})
	}

	sortLayers(added)
	sortLayers(removed)

	return added, removed
}

func resolveTotalSize(analysis ImageAnalysis, result analysisResult) int64 {
	if analysis.TotalSizeBytes != nil {
		return *analysis.TotalSizeBytes
	}
	if result.TotalSizeBytes != nil {
		return *result.TotalSizeBytes
	}
	return 0
}
