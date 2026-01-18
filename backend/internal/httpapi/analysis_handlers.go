package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
)

type AnalysesHandler struct {
	service *analyses.Service
}

func NewAnalysesHandler(service *analyses.Service) *AnalysesHandler {
	return &AnalysesHandler{service: service}
}

type analysisRequest struct {
	RegistryID string `json:"registry_id"`
	Image      string `json:"image"`
	Tag        string `json:"tag"`
}

type analysisResponse struct {
	ID             string          `json:"id"`
	ProjectID      string          `json:"project_id"`
	RegistryID     *string         `json:"registry_id"`
	Image          string          `json:"image"`
	Tag            string          `json:"tag"`
	Status         string          `json:"status"`
	TotalSizeBytes *int64          `json:"total_size_bytes"`
	ResultJSON     json.RawMessage `json:"result_json,omitempty"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
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

	analysis, err := h.service.CreateAnalysis(r.Context(), user.ID, projectID, registryID, req.Image, req.Tag)
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

func toAnalysisResponse(analysis analyses.ImageAnalysis) analysisResponse {
	var registryID *string
	if analysis.RegistryID != nil {
		value := analysis.RegistryID.String()
		registryID = &value
	}

	return analysisResponse{
		ID:             analysis.ID.String(),
		ProjectID:      analysis.ProjectID.String(),
		RegistryID:     registryID,
		Image:          analysis.Image,
		Tag:            analysis.Tag,
		Status:         analysis.Status,
		TotalSizeBytes: analysis.TotalSizeBytes,
		ResultJSON:     analysis.ResultJSON,
		StartedAt:      analysis.StartedAt,
		FinishedAt:     analysis.FinishedAt,
		CreatedAt:      analysis.CreatedAt,
		UpdatedAt:      analysis.UpdatedAt,
	}
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}
