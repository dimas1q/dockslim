package ci

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/budgets"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
	"strings"
)

type compareRepoStub struct {
	analyses map[uuid.UUID]analyses.ImageAnalysis
}

func (r *compareRepoStub) ListAnalysesByProject(ctx context.Context, projectID uuid.UUID) ([]analyses.ImageAnalysis, error) {
	return nil, nil
}

func (r *compareRepoStub) ListHistory(ctx context.Context, projectID uuid.UUID, filter analyses.HistoryFilter) ([]analyses.HistoryItem, error) {
	return nil, nil
}

func (r *compareRepoStub) ListTrends(ctx context.Context, projectID uuid.UUID, metric analyses.TrendMetric, filter analyses.HistoryFilter) ([]analyses.TrendPoint, error) {
	return nil, nil
}

func (r *compareRepoStub) GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	if a, ok := r.analyses[analysisID]; ok {
		return a, nil
	}
	return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
}

func (r *compareRepoStub) GetAnalysisByID(ctx context.Context, analysisID uuid.UUID) (analyses.ImageAnalysis, error) {
	if a, ok := r.analyses[analysisID]; ok {
		return a, nil
	}
	return analyses.ImageAnalysis{}, analyses.ErrAnalysisNotFound
}

func (r *compareRepoStub) GetLatestCompletedBaseline(ctx context.Context, projectID uuid.UUID, image, gitRef string, excludeID uuid.UUID) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, analyses.ErrBaselineNotFound
}

func (r *compareRepoStub) GetProjectPolicy(ctx context.Context, projectID uuid.UUID) (analyses.ProjectPolicy, error) {
	return analyses.ProjectPolicy{}, nil
}

func (r *compareRepoStub) CreateAnalysis(ctx context.Context, params analyses.CreateAnalysisParams) (analyses.ImageAnalysis, error) {
	return analyses.ImageAnalysis{}, nil
}

func (r *compareRepoStub) DeleteAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

func (r *compareRepoStub) RerunAnalysis(ctx context.Context, projectID, analysisID uuid.UUID) error {
	return nil
}

type compareMemberStub struct{}

func (m *compareMemberStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	return "owner", nil
}

type compareRegistryStub struct{}

func (r *compareRegistryStub) GetRegistryForProject(ctx context.Context, projectID, registryID uuid.UUID) (registries.Registry, error) {
	return registries.Registry{ID: registryID, ProjectID: projectID, RegistryURL: "https://registry.example.com"}, nil
}

type budgetResolverStub struct{}

func (b *budgetResolverStub) ResolveBudget(ctx context.Context, userID, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error) {
	return nil, nil
}

func (b *budgetResolverStub) ResolveBudgetForProject(ctx context.Context, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error) {
	return nil, nil
}

// minimal registry struct for interface compliance
type registryStub struct{}

func (r *registryStub) GetRegistryForProject(ctx context.Context, projectID, registryID uuid.UUID) (registries.Registry, error) {
	return registries.Registry{ID: registryID, ProjectID: projectID, RegistryURL: "https://registry.example.com"}, nil
}

func TestCompareReportBuildsMarkdownAndJSON(t *testing.T) {
	projectID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()

	resultJSON, _ := json.Marshal(map[string]any{
		"total_size_bytes": 2048,
		"insights": map[string]any{
			"warnings": []string{"Image is larger than 1GB", "Image has very large layers (>200MB)"},
		},
		"recommendations": []map[string]string{
			{"title": "Use alpine", "severity": "warning", "suggested_action": "Switch base image to alpine"},
			{"title": "Clean cache", "severity": "info", "suggested_action": "Remove package cache"},
		},
	})

	repo := &compareRepoStub{
		analyses: map[uuid.UUID]analyses.ImageAnalysis{
			fromID: {
				ID:         fromID,
				ProjectID:  projectID,
				Image:      "repo/app",
				Tag:        "v1",
				Status:     analyses.StatusCompleted,
				ResultJSON: resultJSON,
			},
			toID: {
				ID:         toID,
				ProjectID:  projectID,
				Image:      "repo/app",
				Tag:        "v2",
				Status:     analyses.StatusCompleted,
				ResultJSON: resultJSON,
			},
		},
	}
	members := &compareMemberStub{}
	regs := &registryStub{}
	analysisSvc := analyses.NewService(repo, members, regs, &budgetResolverStub{})
	ciSvc := NewService(analysisSvc, &budgets.Service{})

	report, err := ciSvc.Compare(context.Background(), projectID, CompareInput{
		FromAnalysisID:  fromID,
		ToAnalysisID:    toID,
		IncludeMarkdown: true,
		IncludeJSON:     true,
	})
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}

	if report.ReportMarkdown == nil || !strings.Contains(*report.ReportMarkdown, "DockSlim Report") {
		t.Fatalf("expected markdown to be generated")
	}
	if report.ReportJSON == nil || report.ReportJSON["delta_bytes"] == nil {
		t.Fatalf("expected json report with delta_bytes")
	}
	if len(report.Warnings) == 0 || len(report.Recommendations) == 0 {
		t.Fatalf("expected highlights extracted from result json")
	}
}

func TestGitHubAndGitLabClients(t *testing.T) {
	projectID := uuid.New()
	toID := uuid.New()

	origGitHub := githubAPI
	origGitLab := gitlabAPI
	defer func() {
		githubAPI = origGitHub
		gitlabAPI = origGitLab
	}()

	githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" || r.Method != http.MethodPost {
			t.Fatalf("github: expected auth header and POST")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer githubServer.Close()

	gitlabServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") == "" || r.Method != http.MethodPost {
			t.Fatalf("gitlab: expected private token and POST")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer gitlabServer.Close()

	analysisSvc := analyses.NewService(&compareRepoStub{}, &compareMemberStub{}, &registryStub{}, &budgetResolverStub{})
	ciSvc := NewService(analysisSvc, &budgets.Service{})
	ciSvc.client = &http.Client{Timeout: 2 * time.Second}
	// override endpoints
	githubAPI = githubServer.URL
	gitlabAPI = gitlabServer.URL

	pr := 1
	if err := ciSvc.PostComment(context.Background(), CommentInput{
		Provider:     "github",
		Repo:         "org/repo",
		PRNumber:     &pr,
		SCMToken:     "ghs_test",
		BodyMarkdown: "test",
		ProjectID:    projectID,
		ToAnalysisID: toID,
	}); err != nil {
		t.Fatalf("github comment failed: %v", err)
	}

	mr := 2
	if err := ciSvc.PostComment(context.Background(), CommentInput{
		Provider:     "gitlab",
		Repo:         "org/proj",
		MRIID:        &mr,
		SCMToken:     "glpat",
		BodyMarkdown: "test",
		ProjectID:    projectID,
		ToAnalysisID: toID,
	}); err != nil {
		t.Fatalf("gitlab comment failed: %v", err)
	}
}

func TestExtractHighlightsSupportsObjectWarnings(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"insights": map[string]any{
			"warnings": []map[string]any{
				{"message": "obj warning"},
				{"text": "text warning"},
			},
		},
		"recommendations": []map[string]string{
			{"title": "Use alpine", "severity": "warning", "suggested_action": "Switch"},
		},
	})
	w, recs := extractHighlights(raw)
	if len(w) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(w))
	}
	if len(recs) != 1 {
		t.Fatalf("expected recommendations")
	}
}

func TestFormatBytesUsesBinaryUnits(t *testing.T) {
	tests := []struct {
		value    int64
		expected string
	}{
		{2048, "2.0 KiB"},
		{134_217_728, "128.0 MiB"},
	}
	for _, tc := range tests {
		if got := formatBytes(tc.value); got != tc.expected {
			t.Fatalf("formatBytes(%d) = %s, want %s", tc.value, got, tc.expected)
		}
	}
	if got := formatSignedBytes(134_217_728); got != "+128.0 MiB" {
		t.Fatalf("formatSignedBytes mismatch, got %s", got)
	}
	if got := formatSignedBytes(-2048); got != "-2.0 KiB" {
		t.Fatalf("formatSignedBytes negative mismatch, got %s", got)
	}
}
