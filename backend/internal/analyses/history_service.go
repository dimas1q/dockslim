package analyses

import (
	"context"
	"errors"

	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

func (s *Service) ListHistory(ctx context.Context, userID, projectID uuid.UUID, filter HistoryFilter) ([]HistoryItem, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return s.repo.ListHistory(ctx, projectID, filter)
}

func (s *Service) ListTrends(ctx context.Context, userID, projectID uuid.UUID, metric TrendMetric, filter HistoryFilter) ([]TrendPoint, error) {
	_, err := s.members.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return s.repo.ListTrends(ctx, projectID, metric, filter)
}

func (s *Service) BaselineCompare(ctx context.Context, userID, analysisID uuid.UUID) (BaselineComparison, error) {
	analysis, err := s.repo.GetAnalysisByID(ctx, analysisID)
	if err != nil {
		return BaselineComparison{}, err
	}

	_, err = s.members.GetMemberRole(ctx, analysis.ProjectID, userID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectMemberNotFound) {
			return BaselineComparison{}, ErrProjectNotFound
		}
		return BaselineComparison{}, err
	}

	if analysis.Status != StatusCompleted {
		return BaselineComparison{}, ErrAnalysisNotCompleted
	}

	policy, err := s.repo.GetProjectPolicy(ctx, analysis.ProjectID)
	if err != nil {
		return BaselineComparison{}, err
	}

	refBranch := policy.BaselineRefBranch
	if refBranch == "" {
		refBranch = DefaultBaselineRef
	}
	mode := policy.BaselineMode
	if mode == "" {
		mode = BaselineModeMainLatest
	}
	if mode != BaselineModeMainLatest {
		refBranch = DefaultBaselineRef
	}

	baseline, err := s.repo.GetLatestCompletedBaseline(ctx, analysis.ProjectID, analysis.Image, refBranch, analysis.ID)
	if err != nil {
		return BaselineComparison{}, err
	}

	currentMetrics, err := resolveMetrics(analysis)
	if err != nil {
		return BaselineComparison{}, err
	}
	if !currentMetrics.hasTotal || !currentMetrics.hasLayerCount {
		return BaselineComparison{}, ErrBaselineMetricsUnavailable
	}
	baselineMetrics, err := resolveMetrics(baseline)
	if err != nil {
		return BaselineComparison{}, err
	}
	if !baselineMetrics.hasTotal || !baselineMetrics.hasLayerCount {
		return BaselineComparison{}, ErrBaselineMetricsUnavailable
	}

	baselineSummary := buildAnalysisSummary(baseline, baselineMetrics)
	baselineSummary.Mode = mode
	baselineSummary.RefBranch = refBranch

	comparison := BaselineComparison{
		AnalysisID: analysis.ID,
		Baseline:   baselineSummary,
		Deltas: BaselineDeltas{
			TotalSizeBytes:    currentMetrics.TotalSizeBytes - baselineMetrics.TotalSizeBytes,
			LayerCount:        currentMetrics.LayerCount - baselineMetrics.LayerCount,
			LargestLayerBytes: currentMetrics.LargestLayerBytes - baselineMetrics.LargestLayerBytes,
		},
		Status: EvaluateBaselineStatus(currentMetrics, baselineMetrics, policy.Thresholds),
	}

	return comparison, nil
}
