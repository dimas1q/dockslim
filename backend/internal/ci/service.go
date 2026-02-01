package ci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/budgets"
	"github.com/google/uuid"
)

const (
	userAgent      = "DockSlim-CI/1.0"
	defaultTimeout = 10 * time.Second
	maxListItems   = 5
)

var (
	githubAPI = "https://api.github.com"
	gitlabAPI = "https://gitlab.com/api/v4"
)

type Service struct {
	analyses *analyses.Service
	budgets  *budgets.Service
	client   *http.Client
	nowFn    func() time.Time
}

func NewService(analyses *analyses.Service, budgets *budgets.Service) *Service {
	return &Service{
		analyses: analyses,
		budgets:  budgets,
		client:   &http.Client{Timeout: defaultTimeout},
		nowFn:    time.Now,
	}
}

type CreateAnalysisInput struct {
	RegistryID uuid.UUID
	Image      string
	Tag        string
}

type CompareInput struct {
	FromAnalysisID  uuid.UUID
	ToAnalysisID    uuid.UUID
	IncludeMarkdown bool
	IncludeJSON     bool
	UIBaseURL       *string
}

type Report struct {
	Comparison      analyses.Comparison `json:"comparison"`
	ReportMarkdown  *string             `json:"report_markdown,omitempty"`
	ReportJSON      map[string]any      `json:"report_json,omitempty"`
	Warnings        []string            `json:"warnings,omitempty"`
	Recommendations []Recommendation    `json:"recommendations,omitempty"`
}

type Recommendation struct {
	Title           string `json:"title"`
	Severity        string `json:"severity"`
	SuggestedAction string `json:"suggested_action"`
}

func (s *Service) CreateAnalysis(ctx context.Context, projectID uuid.UUID, input CreateAnalysisInput) (analyses.ImageAnalysis, error) {
	return s.analyses.CreateAnalysisForCI(ctx, projectID, input.RegistryID, input.Image, input.Tag)
}

func (s *Service) Compare(ctx context.Context, projectID uuid.UUID, input CompareInput) (Report, error) {
	comp, err := s.analyses.CompareAnalysesForProject(ctx, projectID, input.FromAnalysisID, input.ToAnalysisID)
	if err != nil {
		return Report{}, err
	}

	toAnalysis, err := s.analyses.GetAnalysisForProject(ctx, projectID, input.ToAnalysisID)
	if err != nil {
		return Report{}, err
	}

	warnings, recs := extractHighlights(toAnalysis.ResultJSON)

	report := Report{
		Comparison:      comp,
		Warnings:        warnings,
		Recommendations: recs,
	}

	if input.IncludeMarkdown {
		md := buildMarkdown(comp, warnings, recs, input.UIBaseURL)
		report.ReportMarkdown = &md
	}
	if input.IncludeJSON {
		report.ReportJSON = buildReportJSON(comp, warnings, recs)
	}

	return report, nil
}

type CommentInput struct {
	Provider     string
	Repo         string
	PRNumber     *int
	MRIID        *int
	SCMToken     string
	BodyMarkdown string
	ProjectID    uuid.UUID
	ToAnalysisID uuid.UUID
}

func (s *Service) PostComment(ctx context.Context, in CommentInput) error {
	if in.BodyMarkdown == "" {
		return fmt.Errorf("body_markdown is required")
	}

	if len(in.BodyMarkdown) > 20*1024 {
		return fmt.Errorf("body exceeds 20KB limit")
	}

	marker := fmt.Sprintf("<!-- dockslim:project=%s:to=%s -->", in.ProjectID.String(), in.ToAnalysisID.String())
	body := in.BodyMarkdown
	if !strings.Contains(body, marker) {
		body = fmt.Sprintf("%s\n\n%s", body, marker)
	}

	switch strings.ToLower(in.Provider) {
	case "github":
		if in.PRNumber == nil {
			return fmt.Errorf("pr_number is required for github")
		}
		return s.postGitHubComment(ctx, in.Repo, *in.PRNumber, in.SCMToken, body)
	case "gitlab":
		if in.MRIID == nil {
			return fmt.Errorf("mr_iid is required for gitlab")
		}
		return s.postGitLabComment(ctx, in.Repo, *in.MRIID, in.SCMToken, body)
	default:
		return fmt.Errorf("unsupported provider")
	}
}

func (s *Service) postGitHubComment(ctx context.Context, repo string, prNumber int, token, body string) error {
	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", githubAPI, repo, prNumber)
	payload := map[string]string{"body": body}
	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := readBodyLimited(resp.Body, 4096)
		return fmt.Errorf("github api returned status %d: %s", resp.StatusCode, body)
	}
	return nil
}

func (s *Service) postGitLabComment(ctx context.Context, repo string, mrIID int, token, body string) error {
	projectID := url.PathEscape(repo)
	url := fmt.Sprintf("%s/projects/%s/merge_requests/%d/notes", gitlabAPI, projectID, mrIID)
	payload := map[string]string{"body": body}
	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := readBodyLimited(resp.Body, 4096)
		return fmt.Errorf("gitlab api returned status %d: %s", resp.StatusCode, body)
	}
	return nil
}

func readBodyLimited(body io.Reader, limit int64) string {
	if limit <= 0 {
		limit = 4096
	}
	data, err := io.ReadAll(io.LimitReader(body, limit))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (s *Service) httpClient() *http.Client {
	if s.client != nil {
		return s.client
	}
	s.client = &http.Client{Timeout: defaultTimeout}
	return s.client
}

func extractHighlights(raw json.RawMessage) ([]string, []Recommendation) {
	if len(raw) == 0 {
		return nil, nil
	}
	var payload struct {
		Insights struct {
			Warnings any `json:"warnings"`
		} `json:"insights"`
		Recommendations []struct {
			Title           string `json:"title"`
			Severity        string `json:"severity"`
			SuggestedAction string `json:"suggested_action"`
		} `json:"recommendations"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil
	}

	warnings := normalizeWarnings(payload.Insights.Warnings)
	if len(warnings) > maxListItems {
		warnings = warnings[:maxListItems]
	}

	recs := make([]Recommendation, 0, len(payload.Recommendations))
	for i, r := range payload.Recommendations {
		if i >= maxListItems {
			break
		}
		recs = append(recs, Recommendation{
			Title:           r.Title,
			Severity:        r.Severity,
			SuggestedAction: r.SuggestedAction,
		})
	}
	return warnings, recs
}

func normalizeWarnings(value any) []string {
	switch v := value.(type) {
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			switch w := item.(type) {
			case string:
				result = append(result, w)
			case map[string]any:
				if msg, ok := w["message"].(string); ok && msg != "" {
					result = append(result, msg)
					continue
				}
				if msg, ok := w["warning"].(string); ok && msg != "" {
					result = append(result, msg)
					continue
				}
				if msg, ok := w["text"].(string); ok && msg != "" {
					result = append(result, msg)
					continue
				}
			}
		}
		return result
	case []string:
		return v
	default:
		return nil
	}
}

func buildReportJSON(comp analyses.Comparison, warnings []string, recs []Recommendation) map[string]any {
	return map[string]any{
		"from_size_bytes": comp.From.TotalSizeBytes,
		"to_size_bytes":   comp.To.TotalSizeBytes,
		"delta_bytes":     comp.Summary.TotalSizeDiffBytes,
		"budget":          comp.Budget,
		"warnings":        warnings,
		"recommendations": recs,
	}
}

func buildMarkdown(comp analyses.Comparison, warnings []string, recs []Recommendation, uiBaseURL *string) string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "## DockSlim Report for %s\n\n", comp.Image)
	fmt.Fprintf(builder, "| From size | To size | Delta | Impact | Budget |\n")
	fmt.Fprintf(builder, "| --- | --- | --- | --- | --- |\n")
	impact := impactLabel(comp.Summary.TotalSizeDiffBytes)
	budget := "n/a"
	if comp.Budget != nil {
		budget = comp.Budget.Status
	}
	fmt.Fprintf(builder, "| %s | %s | %s | %s | %s |\n",
		formatBytes(comp.From.TotalSizeBytes),
		formatBytes(comp.To.TotalSizeBytes),
		formatSignedBytes(comp.Summary.TotalSizeDiffBytes),
		impact,
		strings.ToUpper(budget),
	)

	if len(warnings) > 0 {
		fmt.Fprintf(builder, "\n**Top Warnings**\n")
		for _, w := range warnings {
			fmt.Fprintf(builder, "- %s\n", w)
		}
	}

	if len(recs) > 0 {
		fmt.Fprintf(builder, "\n**Top Recommendations**\n")
		for _, r := range recs {
			fmt.Fprintf(builder, "- (%s) %s — %s\n", strings.ToUpper(r.Severity), r.Title, r.SuggestedAction)
		}
	}

	if uiBaseURL != nil && *uiBaseURL != "" {
		base := strings.TrimRight(*uiBaseURL, "/")
		fmt.Fprintf(builder, "\n[View analysis](%s/projects/%s/analyses/%s)\n", base, comp.ProjectID.String(), comp.To.AnalysisID.String())
	}

	return builder.String()
}

func impactLabel(delta int64) string {
	switch {
	case delta > 0:
		return "Regression"
	case delta < 0:
		return "Improvement"
	default:
		return "No change"
	}
}

func formatBytes(v int64) string {
	const unit = 1024.0
	val := float64(v)
	if val < unit {
		return fmt.Sprintf("%d B", v)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	val /= unit
	exp := 0
	for val >= unit && exp < len(units)-1 {
		val /= unit
		exp++
	}
	return fmt.Sprintf("%.1f %s", val, units[exp])
}

func formatSignedBytes(v int64) string {
	sign := ""
	if v > 0 {
		sign = "+"
	} else if v < 0 {
		sign = "-"
	}
	abs := v
	if abs < 0 {
		// handle MinInt64 safely
		if abs == math.MinInt64 {
			abs = math.MaxInt64
		} else {
			abs = -abs
		}
	}
	return sign + formatBytes(abs)
}
