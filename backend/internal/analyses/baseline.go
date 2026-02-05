package analyses

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	BaselineStatusOK   = "OK"
	BaselineStatusWarn = "WARN"
	BaselineStatusFail = "FAIL"
)

type BaselineComparison struct {
	AnalysisID uuid.UUID       `json:"analysis_id"`
	Baseline   AnalysisSummary `json:"baseline"`
	Deltas     BaselineDeltas  `json:"deltas"`
	Status     string          `json:"status"`
}

type AnalysisSummary struct {
	AnalysisID        uuid.UUID  `json:"analysis_id"`
	Image             string     `json:"image"`
	Tag               string     `json:"tag"`
	GitRef            *string    `json:"git_ref,omitempty"`
	CommitSHA         *string    `json:"commit_sha,omitempty"`
	AnalyzedAt        *time.Time `json:"analyzed_at,omitempty"`
	TotalSizeBytes    *int64     `json:"total_size_bytes,omitempty"`
	LayerCount        *int       `json:"layer_count,omitempty"`
	LargestLayerBytes *int64     `json:"largest_layer_bytes,omitempty"`
	Mode              string     `json:"mode,omitempty"`
	RefBranch         string     `json:"ref_branch,omitempty"`
}

type BaselineDeltas struct {
	TotalSizeBytes    int64 `json:"total_size_bytes"`
	LayerCount        int   `json:"layer_count"`
	LargestLayerBytes int64 `json:"largest_layer_bytes"`
}

type analysisMetrics struct {
	TotalSizeBytes    int64
	LayerCount        int
	LargestLayerBytes int64
	hasTotal          bool
	hasLayerCount     bool
	hasLargest        bool
}

type analysisResultMeta struct {
	Layers         json.RawMessage `json:"layers"`
	TotalSizeBytes *int64          `json:"total_size_bytes"`
}

func resolveMetrics(analysis ImageAnalysis) (analysisMetrics, error) {
	metrics := analysisMetrics{}
	if analysis.TotalSizeBytes != nil {
		metrics.TotalSizeBytes = *analysis.TotalSizeBytes
		metrics.hasTotal = true
	}
	if analysis.LayerCount != nil {
		metrics.LayerCount = *analysis.LayerCount
		metrics.hasLayerCount = true
	}
	if analysis.LargestLayerBytes != nil {
		metrics.LargestLayerBytes = *analysis.LargestLayerBytes
		metrics.hasLargest = true
	}

	if metrics.hasTotal && metrics.hasLayerCount && metrics.hasLargest {
		return metrics, nil
	}

	var meta analysisResultMeta
	if len(analysis.ResultJSON) == 0 {
		return metrics, nil
	}
	if err := json.Unmarshal(analysis.ResultJSON, &meta); err != nil {
		return metrics, err
	}

	if !metrics.hasTotal && meta.TotalSizeBytes != nil {
		metrics.TotalSizeBytes = *meta.TotalSizeBytes
		metrics.hasTotal = true
	}

	if !metrics.hasLayerCount && meta.Layers != nil {
		var layers []LayerDiff
		if err := json.Unmarshal(meta.Layers, &layers); err != nil {
			return metrics, err
		}
		normalized := normalizeLayers(layers)
		metrics.LayerCount = len(normalized)
		metrics.hasLayerCount = true

		if !metrics.hasLargest {
			var max int64
			for _, layer := range normalized {
				if layer.SizeBytes > max {
					max = layer.SizeBytes
				}
			}
			metrics.LargestLayerBytes = max
			metrics.hasLargest = true
		}
	}

	if metrics.hasLargest {
		return metrics, nil
	}

	result, err := parseAnalysisResult(analysis.ResultJSON)
	if err != nil {
		return metrics, err
	}

	if !metrics.hasLargest {
		layers := normalizeLayers(result.Layers)
		var max int64
		for _, layer := range layers {
			if layer.SizeBytes > max {
				max = layer.SizeBytes
			}
		}
		metrics.LargestLayerBytes = max
		metrics.hasLargest = true
	}

	return metrics, nil
}

func buildAnalysisSummary(analysis ImageAnalysis, metrics analysisMetrics) AnalysisSummary {
	var total *int64
	var layerCount *int
	var largest *int64

	if metrics.hasTotal {
		value := metrics.TotalSizeBytes
		total = &value
	}
	if metrics.hasLayerCount {
		value := metrics.LayerCount
		layerCount = &value
	}
	if metrics.hasLargest {
		value := metrics.LargestLayerBytes
		largest = &value
	}

	return AnalysisSummary{
		AnalysisID:        analysis.ID,
		Image:             analysis.Image,
		Tag:               analysis.Tag,
		GitRef:            analysis.GitRef,
		CommitSHA:         analysis.CommitSHA,
		AnalyzedAt:        analysis.AnalyzedAt,
		TotalSizeBytes:    total,
		LayerCount:        layerCount,
		LargestLayerBytes: largest,
	}
}

func EvaluateBaselineStatus(current, baseline analysisMetrics, thresholds Thresholds) string {
	status := BaselineStatusOK

	deltaBytes := current.TotalSizeBytes - baseline.TotalSizeBytes
	deltaLayers := current.LayerCount - baseline.LayerCount

	if thresholds.HardLimitBytes != nil && current.TotalSizeBytes > *thresholds.HardLimitBytes {
		return BaselineStatusFail
	}

	if deltaBytes > 0 {
		if thresholds.FailDeltaBytes != nil && deltaBytes > *thresholds.FailDeltaBytes {
			return BaselineStatusFail
		}
		if thresholds.WarnDeltaBytes != nil && deltaBytes > *thresholds.WarnDeltaBytes {
			status = BaselineStatusWarn
		}
	}

	if deltaLayers > 0 {
		if thresholds.FailDeltaLayers != nil && deltaLayers > *thresholds.FailDeltaLayers {
			return BaselineStatusFail
		}
		if thresholds.WarnDeltaLayers != nil && deltaLayers > *thresholds.WarnDeltaLayers {
			if status == BaselineStatusOK {
				status = BaselineStatusWarn
			}
		}
	}

	return status
}
