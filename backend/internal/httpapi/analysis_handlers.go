package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/featureflags"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

type AnalysesHandler struct {
	service      *analyses.Service
	featureFlags FeatureGate
}

type FeatureGate interface {
	GetUserFeatures(ctx context.Context, userID uuid.UUID) (featureflags.UserFeatures, error)
	HasFeature(ctx context.Context, userID uuid.UUID, featureName string) (bool, error)
}

func NewAnalysesHandler(service *analyses.Service, featureFlags ...FeatureGate) *AnalysesHandler {
	handler := &AnalysesHandler{service: service}
	if len(featureFlags) > 0 {
		handler.featureFlags = featureFlags[0]
	}
	return handler
}

type analysisRequest struct {
	RegistryID string  `json:"registry_id"`
	Image      string  `json:"image"`
	Tag        string  `json:"tag"`
	GitRef     *string `json:"git_ref,omitempty"`
	CommitSHA  *string `json:"commit_sha,omitempty"`
}

type analysisResponse struct {
	ID                string          `json:"id"`
	ProjectID         string          `json:"project_id"`
	RegistryID        *string         `json:"registry_id"`
	Image             string          `json:"image"`
	Tag               string          `json:"tag"`
	GitRef            *string         `json:"git_ref,omitempty"`
	CommitSHA         *string         `json:"commit_sha,omitempty"`
	Status            string          `json:"status"`
	TotalSizeBytes    *int64          `json:"total_size_bytes"`
	LayerCount        *int            `json:"layer_count,omitempty"`
	LargestLayerBytes *int64          `json:"largest_layer_bytes,omitempty"`
	ResultJSON        json.RawMessage `json:"result_json,omitempty"`
	StartedAt         *time.Time      `json:"started_at,omitempty"`
	FinishedAt        *time.Time      `json:"finished_at,omitempty"`
	AnalyzedAt        *time.Time      `json:"analyzed_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type historyResponse struct {
	ID                string     `json:"id"`
	Image             string     `json:"image"`
	GitRef            *string    `json:"git_ref,omitempty"`
	CommitSHA         *string    `json:"commit_sha,omitempty"`
	Status            string     `json:"status"`
	AnalyzedAt        *time.Time `json:"analyzed_at,omitempty"`
	TotalSizeBytes    *int64     `json:"total_size_bytes,omitempty"`
	LayerCount        *int       `json:"layer_count,omitempty"`
	LargestLayerBytes *int64     `json:"largest_layer_bytes,omitempty"`
}

type trendResponse struct {
	Ts    time.Time `json:"ts"`
	Value int64     `json:"value"`
}

type baselineCompareResponse struct {
	AnalysisID string                  `json:"analysis_id"`
	Baseline   baselineSummaryResponse `json:"baseline"`
	Deltas     baselineDeltasResponse  `json:"deltas"`
	Status     string                  `json:"status"`
}

type baselineSummaryResponse struct {
	AnalysisID        string     `json:"analysis_id"`
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

type baselineDeltasResponse struct {
	TotalSizeBytes    int64 `json:"total_size_bytes"`
	LayerCount        int   `json:"layer_count"`
	LargestLayerBytes int64 `json:"largest_layer_bytes"`
}

func (h *AnalysesHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	items, err := h.service.ListAnalyses(r.Context(), user.ID, projectID)
	if err != nil {
		if errors.Is(err, analyses.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to list analyses")
		return
	}

	resp := make([]analysisResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toAnalysisResponse(item))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AnalysesHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req analysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RegistryID == "" {
		writeError(w, http.StatusBadRequest, "registry_id is required")
		return
	}
	registryID, err := parseUUID(req.RegistryID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid registry id")
		return
	}

	analysis, err := h.service.CreateAnalysis(r.Context(), user.ID, projectID, registryID, req.Image, req.Tag, req.GitRef, req.CommitSHA)
	if err != nil {
		switch {
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, analyses.ErrInvalidImage),
			errors.Is(err, analyses.ErrInvalidRegistry),
			errors.Is(err, analyses.ErrRegistryMismatch):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, registries.ErrRegistryNotFound):
			writeError(w, http.StatusNotFound, "registry not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to create analysis")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toAnalysisResponse(analysis))
}

func (h *AnalysesHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	analysisID, err := parseUUIDParam(r, "analysisId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	analysis, err := h.service.GetAnalysis(r.Context(), user.ID, projectID, analysisID)
	if err != nil {
		switch {
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch analysis")
		}
		return
	}

	if h.featureFlags != nil && len(analysis.ResultJSON) > 0 {
		advancedEnabled, err := h.featureFlags.HasFeature(r.Context(), user.ID, featureflags.FeatureAdvancedInsights)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		if !advancedEnabled {
			analysis.ResultJSON = stripAdvancedInsights(analysis.ResultJSON)
		}
	}

	writeJSON(w, http.StatusOK, toAnalysisResponse(analysis))
}

func (h *AnalysesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	analysisID, err := parseUUIDParam(r, "analysisId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	if err := h.service.DeleteAnalysis(r.Context(), user.ID, projectID, analysisID); err != nil {
		switch {
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete analysis")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AnalysesHandler) Rerun(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	analysisID, err := parseUUIDParam(r, "analysisId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	if err := h.service.RerunAnalysis(r.Context(), user.ID, projectID, analysisID); err != nil {
		switch {
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		case errors.Is(err, analyses.ErrAnalysisRunning):
			writeError(w, http.StatusConflict, "analysis is running")
		default:
			writeError(w, http.StatusInternalServerError, "failed to rerun analysis")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AnalysesHandler) Compare(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	fromID, err := parseUUID(r.URL.Query().Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from analysis id")
		return
	}
	toID, err := parseUUID(r.URL.Query().Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to analysis id")
		return
	}

	comparison, err := h.service.CompareAnalyses(r.Context(), user.ID, projectID, fromID, toID)
	if err != nil {
		switch {
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		case errors.Is(err, analyses.ErrAnalysesDifferentImage):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, analyses.ErrAnalysesNotCompleted):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to compare analyses")
		}
		return
	}

	writeJSON(w, http.StatusOK, comparison)
}

func (h *AnalysesHandler) History(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	filter, err := parseHistoryFilter(r, 100)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.featureFlags != nil {
		featureSet, err := h.featureFlags.GetUserFeatures(r.Context(), user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		applyHistoryDaysLimit(&filter, featureSet.Features)
	}

	items, err := h.service.ListHistory(r.Context(), user.ID, projectID, filter)
	if err != nil {
		if errors.Is(err, analyses.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	resp := make([]historyResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, historyResponse{
			ID:                item.ID.String(),
			Image:             item.Image,
			GitRef:            item.GitRef,
			CommitSHA:         item.CommitSHA,
			Status:            item.Status,
			AnalyzedAt:        item.AnalyzedAt,
			TotalSizeBytes:    item.TotalSizeBytes,
			LayerCount:        item.LayerCount,
			LargestLayerBytes: item.LargestLayerBytes,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AnalysesHandler) Trends(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	metricStr := strings.TrimSpace(r.URL.Query().Get("metric"))
	if metricStr == "" {
		writeError(w, http.StatusBadRequest, "metric is required")
		return
	}
	metric := analyses.TrendMetric(metricStr)
	switch metric {
	case analyses.TrendMetricTotalSize, analyses.TrendMetricLayerCount, analyses.TrendMetricLargestLayer:
	default:
		writeError(w, http.StatusBadRequest, "invalid metric")
		return
	}

	filter, err := parseHistoryFilter(r, 1000)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.featureFlags != nil {
		featureSet, err := h.featureFlags.GetUserFeatures(r.Context(), user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		if metric == analyses.TrendMetricLargestLayer {
			enabled, ok := featureSet.FeatureValue(featureflags.FeatureAdvancedTrends)
			if !ok || !featureflags.FeatureEnabled(enabled) {
				writeError(w, http.StatusForbidden, "feature not available on current plan")
				return
			}
		}
		applyHistoryDaysLimit(&filter, featureSet.Features)
	}

	points, err := h.service.ListTrends(r.Context(), user.ID, projectID, metric, filter)
	if err != nil {
		if errors.Is(err, analyses.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch trends")
		return
	}

	resp := make([]trendResponse, 0, len(points))
	for _, point := range points {
		resp = append(resp, trendResponse{
			Ts:    point.Timestamp,
			Value: point.Value,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AnalysesHandler) BaselineCompare(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if h.featureFlags != nil {
		enabled, err := h.featureFlags.HasFeature(r.Context(), user.ID, featureflags.FeatureBaselineSLA)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		if !enabled {
			writeError(w, http.StatusForbidden, "feature not available on current plan")
			return
		}
	}

	analysisID, err := parseUUIDParam(r, "analysisId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	comparison, err := h.service.BaselineCompare(r.Context(), user.ID, analysisID)
	if err != nil {
		switch {
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrBaselineNotFound):
			writeError(w, http.StatusNotFound, "no baseline analysis found")
		case errors.Is(err, analyses.ErrAnalysisNotCompleted):
			writeError(w, http.StatusConflict, "analysis is not completed")
		case errors.Is(err, analyses.ErrBaselineMetricsUnavailable):
			writeError(w, http.StatusConflict, "baseline metrics unavailable")
		default:
			writeError(w, http.StatusInternalServerError, "failed to compare baseline")
		}
		return
	}

	resp := baselineCompareResponse{
		AnalysisID: comparison.AnalysisID.String(),
		Baseline: baselineSummaryResponse{
			AnalysisID:        comparison.Baseline.AnalysisID.String(),
			Image:             comparison.Baseline.Image,
			Tag:               comparison.Baseline.Tag,
			GitRef:            comparison.Baseline.GitRef,
			CommitSHA:         comparison.Baseline.CommitSHA,
			AnalyzedAt:        comparison.Baseline.AnalyzedAt,
			TotalSizeBytes:    comparison.Baseline.TotalSizeBytes,
			LayerCount:        comparison.Baseline.LayerCount,
			LargestLayerBytes: comparison.Baseline.LargestLayerBytes,
			Mode:              comparison.Baseline.Mode,
			RefBranch:         comparison.Baseline.RefBranch,
		},
		Deltas: baselineDeltasResponse{
			TotalSizeBytes:    comparison.Deltas.TotalSizeBytes,
			LayerCount:        comparison.Deltas.LayerCount,
			LargestLayerBytes: comparison.Deltas.LargestLayerBytes,
		},
		Status: comparison.Status,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AnalysesHandler) ExportJSON(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if h.featureFlags != nil {
		enabled, err := h.featureFlags.HasFeature(r.Context(), user.ID, featureflags.FeatureExportJSON)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		if !enabled {
			writeError(w, http.StatusForbidden, "feature not available on current plan")
			return
		}
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	analysisID, err := parseUUIDParam(r, "analysisId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	analysis, err := h.service.GetAnalysis(r.Context(), user.ID, projectID, analysisID)
	if err != nil {
		switch {
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch analysis")
		}
		return
	}
	if analysis.Status != analyses.StatusCompleted || len(analysis.ResultJSON) == 0 {
		writeError(w, http.StatusConflict, "analysis is not completed")
		return
	}

	resultJSON := analysis.ResultJSON
	if h.featureFlags != nil {
		advancedEnabled, err := h.featureFlags.HasFeature(r.Context(), user.ID, featureflags.FeatureAdvancedInsights)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		if !advancedEnabled {
			resultJSON = stripAdvancedInsights(resultJSON)
		}
	}

	filename := buildAnalysisExportFileName(analysis, "json")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	_, _ = w.Write(resultJSON)
}

func (h *AnalysesHandler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if h.featureFlags != nil {
		enabled, err := h.featureFlags.HasFeature(r.Context(), user.ID, featureflags.FeatureExportPDF)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		if !enabled {
			writeError(w, http.StatusForbidden, "feature not available on current plan")
			return
		}
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	analysisID, err := parseUUIDParam(r, "analysisId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	analysis, err := h.service.GetAnalysis(r.Context(), user.ID, projectID, analysisID)
	if err != nil {
		switch {
		case errors.Is(err, analyses.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch analysis")
		}
		return
	}

	if analysis.Status != analyses.StatusCompleted || len(analysis.ResultJSON) == 0 {
		writeError(w, http.StatusConflict, "analysis is not completed")
		return
	}

	resultJSON := analysis.ResultJSON
	if h.featureFlags != nil {
		advancedEnabled, err := h.featureFlags.HasFeature(r.Context(), user.ID, featureflags.FeatureAdvancedInsights)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to resolve feature access")
			return
		}
		if !advancedEnabled {
			resultJSON = stripAdvancedInsights(resultJSON)
		}
	}

	document, err := buildAnalysisPDF(analysis, resultJSON)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate pdf export")
		return
	}
	filename := buildAnalysisExportFileName(analysis, "pdf")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	_, _ = w.Write(document)
}

var exportFileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func buildAnalysisExportFileName(analysis analyses.ImageAnalysis, ext string) string {
	base := fmt.Sprintf("%s-%s-%s", analysis.Image, analysis.Tag, analysis.ID.String()[:8])
	base = strings.ToLower(base)
	base = exportFileNameSanitizer.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._-")
	if base == "" {
		base = "analysis-export"
	}
	return base + "." + ext
}

func buildAnalysisPDF(analysis analyses.ImageAnalysis, resultJSON json.RawMessage) ([]byte, error) {
	type pdfSummary struct {
		Recommendations []any `json:"recommendations"`
	}
	var summary pdfSummary
	if len(resultJSON) > 0 {
		_ = json.Unmarshal(resultJSON, &summary)
	}

	analyzedAt := "-"
	if analysis.AnalyzedAt != nil {
		analyzedAt = analysis.AnalyzedAt.UTC().Format(time.RFC3339)
	}
	totalSize := "-"
	if analysis.TotalSizeBytes != nil {
		totalSize = strconv.FormatInt(*analysis.TotalSizeBytes, 10)
	}
	layerCount := "-"
	if analysis.LayerCount != nil {
		layerCount = strconv.Itoa(*analysis.LayerCount)
	}
	largestLayer := "-"
	if analysis.LargestLayerBytes != nil {
		largestLayer = strconv.FormatInt(*analysis.LargestLayerBytes, 10)
	}

	rows := [][2]string{
		{"Image", analysis.Image},
		{"Tag", analysis.Tag},
		{"Analysis ID", analysis.ID.String()},
		{"Analyzed At", analyzedAt},
		{"Total Size (bytes)", totalSize},
		{"Layer Count", layerCount},
		{"Largest Layer (bytes)", largestLayer},
		{"Recommendations", strconv.Itoa(len(summary.Recommendations))},
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("DockSlim Analysis Report", false)
	pdf.SetCreator("DockSlim", false)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, safePDFText("DockSlim Analysis Report"), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	for _, row := range rows {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(52, 7, safePDFText(row[0]), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 11)
		pdf.MultiCell(0, 7, safePDFText(row[1]), "", "L", false)
	}

	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 7, safePDFText("Generated At: "+time.Now().UTC().Format(time.RFC3339)), "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func safePDFText(value string) string {
	if value == "" {
		return ""
	}
	var out strings.Builder
	out.Grow(len(value))
	for _, r := range value {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			out.WriteRune(' ')
		case r < 32:
			continue
		case r > 126:
			out.WriteRune('?')
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}

func toAnalysisResponse(analysis analyses.ImageAnalysis) analysisResponse {
	var registryID *string
	if analysis.RegistryID != nil {
		value := analysis.RegistryID.String()
		registryID = &value
	}

	return analysisResponse{
		ID:                analysis.ID.String(),
		ProjectID:         analysis.ProjectID.String(),
		RegistryID:        registryID,
		Image:             analysis.Image,
		Tag:               analysis.Tag,
		GitRef:            analysis.GitRef,
		CommitSHA:         analysis.CommitSHA,
		Status:            analysis.Status,
		TotalSizeBytes:    analysis.TotalSizeBytes,
		LayerCount:        analysis.LayerCount,
		LargestLayerBytes: analysis.LargestLayerBytes,
		ResultJSON:        analysis.ResultJSON,
		StartedAt:         analysis.StartedAt,
		FinishedAt:        analysis.FinishedAt,
		AnalyzedAt:        analysis.AnalyzedAt,
		CreatedAt:         analysis.CreatedAt,
		UpdatedAt:         analysis.UpdatedAt,
	}
}

func stripAdvancedInsights(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return raw
	}

	for key := range payload {
		if isAdvancedFieldKey(key) {
			delete(payload, key)
		}
	}

	if insights, ok := payload["insights"].(map[string]any); ok {
		payload["insights"] = stripAdvancedFromInsights(insights)
	}
	delete(payload, "recommendations")

	out, err := json.Marshal(payload)
	if err != nil {
		return raw
	}
	return out
}

func stripAdvancedFromInsights(insights map[string]any) map[string]any {
	result := make(map[string]any, 1)
	value, ok := insights["warnings"]
	if !ok {
		return result
	}
	warnings, ok := value.([]any)
	if !ok {
		return result
	}
	result["warnings"] = stripAdvancedWarnings(warnings)
	return result
}

func stripAdvancedWarnings(warnings []any) []any {
	out := make([]any, 0, len(warnings))
	for _, warning := range warnings {
		if isAdvancedWarning(warning) {
			continue
		}
		out = append(out, warning)
	}
	return out
}

func isAdvancedFieldKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return strings.HasPrefix(normalized, "advanced_") || strings.HasSuffix(normalized, "_advanced")
}

func isAdvancedWarning(value any) bool {
	switch warning := value.(type) {
	case string:
		normalized := strings.ToLower(strings.TrimSpace(warning))
		return strings.HasPrefix(normalized, "adv:") || strings.HasPrefix(normalized, "[adv]") || strings.HasPrefix(normalized, "advanced:")
	case map[string]any:
		if tier, ok := warning["tier"].(string); ok {
			if isAdvancedTier(tier) {
				return true
			}
		}
		if advanced, ok := warning["advanced"].(bool); ok && advanced {
			return true
		}
		if id, ok := warning["id"].(string); ok {
			return isAdvancedID(id)
		}
	}
	return false
}

func isAdvancedTier(tier string) bool {
	normalized := strings.ToLower(strings.TrimSpace(tier))
	switch normalized {
	case "pro", "team", "advanced", "premium":
		return true
	default:
		return false
	}
}

func isAdvancedID(id string) bool {
	normalized := strings.ToLower(strings.TrimSpace(id))
	return strings.HasPrefix(normalized, "adv_") || strings.HasPrefix(normalized, "advanced_")
}

func applyHistoryDaysLimit(filter *analyses.HistoryFilter, features map[string]any) {
	if filter == nil {
		return
	}
	value, ok := features[featureflags.FeatureHistoryDaysLimit]
	if !ok || value == nil {
		return
	}

	days := intFromAny(value)
	if days <= 0 {
		return
	}
	start := time.Now().UTC().AddDate(0, 0, -days)
	if filter.From == nil || filter.From.Before(start) {
		filter.From = &start
	}
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func parseHistoryFilter(r *http.Request, defaultLimit int) (analyses.HistoryFilter, error) {
	query := r.URL.Query()
	filter := analyses.HistoryFilter{}

	if image := strings.TrimSpace(query.Get("image")); image != "" {
		filter.Image = &image
	}
	if gitRef := strings.TrimSpace(query.Get("git_ref")); gitRef != "" {
		filter.GitRef = &gitRef
	}
	if status := strings.TrimSpace(query.Get("status")); status != "" {
		switch status {
		case analyses.HistoryStatusAll,
			analyses.HistoryStatusQueued,
			analyses.HistoryStatusRunning,
			analyses.HistoryStatusFailed,
			analyses.HistoryStatusComplete:
			filter.Status = status
		default:
			return analyses.HistoryFilter{}, errors.New("invalid status")
		}
	}

	if fromValue := strings.TrimSpace(query.Get("from")); fromValue != "" {
		parsed, err := parseTimeQuery(fromValue, false)
		if err != nil {
			return analyses.HistoryFilter{}, errors.New("invalid from date")
		}
		filter.From = parsed
	}
	if toValue := strings.TrimSpace(query.Get("to")); toValue != "" {
		parsed, err := parseTimeQuery(toValue, true)
		if err != nil {
			return analyses.HistoryFilter{}, errors.New("invalid to date")
		}
		filter.To = parsed
	}

	if limitValue := strings.TrimSpace(query.Get("limit")); limitValue != "" {
		limit, err := strconv.Atoi(limitValue)
		if err != nil || limit <= 0 {
			return analyses.HistoryFilter{}, errors.New("invalid limit")
		}
		if limit > 500 {
			limit = 500
		}
		filter.Limit = limit
	} else if defaultLimit > 0 {
		filter.Limit = defaultLimit
	}

	return filter, nil
}

func parseTimeQuery(value string, isEnd bool) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed, nil
	}
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		if isEnd {
			end := parsed.Add(24*time.Hour - time.Nanosecond)
			return &end, nil
		}
		start := parsed
		return &start, nil
	}
	return nil, errors.New("invalid time")
}
