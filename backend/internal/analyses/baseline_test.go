package analyses

import "testing"

func TestEvaluateBaselineStatus(t *testing.T) {
	warnBytes := int64(50)
	failBytes := int64(150)
	hardLimit := int64(1000)
	warnLayers := 2
	failLayers := 4
	thresholds := Thresholds{
		WarnDeltaBytes:  &warnBytes,
		FailDeltaBytes:  &failBytes,
		HardLimitBytes:  &hardLimit,
		WarnDeltaLayers: &warnLayers,
		FailDeltaLayers: &failLayers,
	}

	tests := []struct {
		name     string
		current  analysisMetrics
		baseline analysisMetrics
		want     string
	}{
		{
			name:     "ok under thresholds",
			current:  analysisMetrics{TotalSizeBytes: 120, LayerCount: 5},
			baseline: analysisMetrics{TotalSizeBytes: 100, LayerCount: 5},
			want:     BaselineStatusOK,
		},
		{
			name:     "warn on size delta",
			current:  analysisMetrics{TotalSizeBytes: 180, LayerCount: 5},
			baseline: analysisMetrics{TotalSizeBytes: 100, LayerCount: 5},
			want:     BaselineStatusWarn,
		},
		{
			name:     "fail on size delta",
			current:  analysisMetrics{TotalSizeBytes: 300, LayerCount: 5},
			baseline: analysisMetrics{TotalSizeBytes: 100, LayerCount: 5},
			want:     BaselineStatusFail,
		},
		{
			name:     "fail on hard limit",
			current:  analysisMetrics{TotalSizeBytes: 1200, LayerCount: 5},
			baseline: analysisMetrics{TotalSizeBytes: 1190, LayerCount: 5},
			want:     BaselineStatusFail,
		},
		{
			name:     "warn on layer delta",
			current:  analysisMetrics{TotalSizeBytes: 120, LayerCount: 8},
			baseline: analysisMetrics{TotalSizeBytes: 100, LayerCount: 5},
			want:     BaselineStatusWarn,
		},
		{
			name:     "fail on layer delta",
			current:  analysisMetrics{TotalSizeBytes: 120, LayerCount: 11},
			baseline: analysisMetrics{TotalSizeBytes: 100, LayerCount: 5},
			want:     BaselineStatusFail,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateBaselineStatus(tc.current, tc.baseline, thresholds)
			if got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}
