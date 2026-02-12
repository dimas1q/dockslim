package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/ci"
	"github.com/dimas1q/dockslim/backend/internal/featureflags"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
)

type ciServiceStub struct {
	report ci.Report
	err    error
}

func (s *ciServiceStub) CreateAnalysis(ctx context.Context, projectID uuid.UUID, input ci.CreateAnalysisInput) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, nil
}

func (s *ciServiceStub) Compare(ctx context.Context, projectID uuid.UUID, input ci.CompareInput) (ci.Report, error) {
	if s.err != nil {
		return ci.Report{}, s.err
	}
	return s.report, nil
}

func (s *ciServiceStub) PostComment(ctx context.Context, in ci.CommentInput) error {
	return s.err
}

type ciAnalysisServiceStub struct{}

func (s *ciAnalysisServiceStub) CreateAnalysis(ctx context.Context, userID, projectID uuid.UUID, registryID uuid.UUID, image, tag string, gitRef, commitSHA *string) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, nil
}

func (s *ciAnalysisServiceStub) GetAnalysis(ctx context.Context, userID, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{
		ID:        analysisID,
		ProjectID: projectID,
		Status:    analyses.StatusCompleted,
	}, nil
}

func (s *ciAnalysisServiceStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{
		ID:        analysisID,
		ProjectID: projectID,
		Status:    analyses.StatusCompleted,
	}, nil
}

func (s *ciAnalysisServiceStub) CompareAnalyses(ctx context.Context, userID, projectID, fromID, toID uuid.UUID) (analyses.Comparison, error) {
	return analyses.Comparison{
		ProjectID: projectID,
		Image:     "repo/app",
	}, nil
}

type ciRegistryResolverStub struct{}

func (s *ciRegistryResolverStub) ResolveRegistryReference(ctx context.Context, projectID uuid.UUID, registryID *uuid.UUID, name *string, host *string) (registries.Registry, error) {
	id := uuid.New()
	if registryID != nil {
		id = *registryID
	}
	return registries.Registry{ID: id, ProjectID: projectID}, nil
}

func TestCIHandlerCompareReportBlocksJSONForFreePlan(t *testing.T) {
	projectID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	user := auth.User{ID: uuid.New()}

	handler := NewCIHandler(&ciServiceStub{}, &ciAnalysisServiceStub{}, &ciRegistryResolverStub{}, &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureExportJSON: false,
				},
			},
		},
	})

	payload := map[string]any{
		"project_id":       projectID.String(),
		"from_analysis_id": fromID.String(),
		"to_analysis_id":   toID.String(),
		"include_markdown": false,
		"include_json":     true,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci/reports/compare", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.CompareReport(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCIHandlerPostCommentLimitedModeEnforcesBodyLimit(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}

	handler := NewCIHandler(&ciServiceStub{}, &ciAnalysisServiceStub{}, &ciRegistryResolverStub{}, &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureCIComments: featureflags.CICommentsModeLimited,
				},
			},
		},
	})

	payload := map[string]any{
		"project_id":     projectID.String(),
		"provider":       "github",
		"repo":           "owner/repo",
		"pr_number":      1,
		"scm_token":      "token",
		"body_markdown":  strings.Repeat("a", 2001),
		"to_analysis_id": analysisID.String(),
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci/comments", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.PostComment(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCIHandlerPostCommentLimitedModeAllowsShortBody(t *testing.T) {
	projectID := uuid.New()
	analysisID := uuid.New()
	user := auth.User{ID: uuid.New()}

	handler := NewCIHandler(&ciServiceStub{}, &ciAnalysisServiceStub{}, &ciRegistryResolverStub{}, &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureCIComments: featureflags.CICommentsModeLimited,
				},
			},
		},
	})

	payload := map[string]any{
		"project_id":     projectID.String(),
		"provider":       "github",
		"repo":           "owner/repo",
		"pr_number":      1,
		"scm_token":      "token",
		"body_markdown":  "ok",
		"to_analysis_id": analysisID.String(),
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci/comments", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.PostComment(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestCIHandlerCompareReportSanitizesRecommendationsForFreePlan(t *testing.T) {
	projectID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	user := auth.User{ID: uuid.New()}

	handler := NewCIHandler(&ciServiceStub{
		report: ci.Report{
			Comparison: analyses.Comparison{
				ProjectID: projectID,
				Image:     "repo/app",
			},
			Warnings: []string{"disk pressure warning"},
			Recommendations: []ci.Recommendation{
				{Title: "Trim layers", Severity: "warning", SuggestedAction: "Use multi-stage builds"},
			},
			ReportJSON: map[string]any{
				"warnings":        []any{"disk pressure warning"},
				"recommendations": []any{map[string]any{"title": "Trim layers"}},
			},
			ReportMarkdown: ptr("## DockSlim Report for repo/app\n\n**Top Warnings**\n- disk pressure warning\n\n**Top Recommendations**\n- (WARNING) Trim layers — Use multi-stage builds\n\n[View analysis](https://example.com)\n"),
		},
	}, &ciAnalysisServiceStub{}, &ciRegistryResolverStub{}, &featureGateStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				Features: map[string]any{
					featureflags.FeatureAdvancedInsights: false,
					featureflags.FeatureExportJSON:       true,
				},
			},
		},
	})

	payload := map[string]any{
		"project_id":       projectID.String(),
		"from_analysis_id": fromID.String(),
		"to_analysis_id":   toID.String(),
		"include_markdown": true,
		"include_json":     true,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ci/reports/compare", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.CompareReport(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response ci.Report
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Warnings) == 0 {
		t.Fatalf("expected core warnings to remain")
	}
	if len(response.Recommendations) != 0 {
		t.Fatalf("expected recommendations to be hidden for free plan")
	}
	if response.ReportJSON != nil {
		if _, hasRecommendations := response.ReportJSON["recommendations"]; hasRecommendations {
			t.Fatalf("expected report_json recommendations to be removed")
		}
	}
	if response.ReportMarkdown != nil && strings.Contains(*response.ReportMarkdown, "Top Recommendations") {
		t.Fatalf("expected markdown recommendations section to be removed")
	}
}

func ptr[T any](v T) *T {
	return &v
}
