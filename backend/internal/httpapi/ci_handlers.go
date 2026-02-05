package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dimas1q/dockslim/backend/internal/analyses"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/ci"
	"github.com/dimas1q/dockslim/backend/internal/citokens"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/google/uuid"
)

type CIHandler struct {
	ciService  CIService
	analyses   AnalysisService
	registries RegistryResolver
}

type CIService interface {
	CreateAnalysis(ctx context.Context, projectID uuid.UUID, input ci.CreateAnalysisInput) (analyses.ImageAnalysis, error)
	Compare(ctx context.Context, projectID uuid.UUID, input ci.CompareInput) (ci.Report, error)
	PostComment(ctx context.Context, in ci.CommentInput) error
}

type AnalysisService interface {
	CreateAnalysis(ctx context.Context, userID, projectID uuid.UUID, registryID uuid.UUID, image, tag string, gitRef, commitSHA *string) (analyses.ImageAnalysis, error)
	GetAnalysis(ctx context.Context, userID, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error)
	GetAnalysisForProject(ctx context.Context, projectID, analysisID uuid.UUID) (analyses.ImageAnalysis, error)
	CompareAnalyses(ctx context.Context, userID, projectID, fromID, toID uuid.UUID) (analyses.Comparison, error)
}

type RegistryResolver interface {
	ResolveRegistryReference(ctx context.Context, projectID uuid.UUID, registryID *uuid.UUID, name *string, host *string) (registries.Registry, error)
}

func NewCIHandler(ciService CIService, analyses AnalysisService, registries RegistryResolver) *CIHandler {
	return &CIHandler{ciService: ciService, analyses: analyses, registries: registries}
}

type ciImageReportRequest struct {
	ProjectID    *string `json:"project_id,omitempty"`
	RegistryID   string  `json:"registry_id,omitempty"`
	RegistryName *string `json:"registry_name,omitempty"`
	RegistryHost *string `json:"registry_host,omitempty"`
	Image        string  `json:"image"`
	Tag          string  `json:"tag"`
	GitRef       *string `json:"git_ref,omitempty"`
	CommitSHA    *string `json:"commit_sha,omitempty"`
}

func (h *CIHandler) CreateAnalysisReport(w http.ResponseWriter, r *http.Request) {
	projectID, user, ciToken, err := resolveProject(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ciImageReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if projectID == nil {
		if req.ProjectID == nil || *req.ProjectID == "" {
			writeError(w, http.StatusBadRequest, "project_id is required")
			return
		}
		parsed, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		projectID = &parsed
	} else if req.ProjectID != nil && *req.ProjectID != "" && *req.ProjectID != projectID.String() {
		writeError(w, http.StatusForbidden, "token does not match project")
		return
	}

	registryID, err := h.resolveRegistry(r.Context(), *projectID, req.RegistryID, req.RegistryName, req.RegistryHost)
	if err != nil {
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, registries.ErrRegistryNotFound):
			status = http.StatusNotFound
		case errors.Is(err, registries.ErrRegistryAmbiguous):
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}

	var analysis analyses.ImageAnalysis
	if ciToken != nil {
		analysis, err = h.ciService.CreateAnalysis(r.Context(), *projectID, ci.CreateAnalysisInput{
			RegistryID: registryID,
			Image:      req.Image,
			Tag:        req.Tag,
			GitRef:     req.GitRef,
			CommitSHA:  req.CommitSHA,
		})
	} else {
		analysis, err = h.analyses.CreateAnalysis(r.Context(), user.ID, *projectID, registryID, req.Image, req.Tag, req.GitRef, req.CommitSHA)
	}

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
		default:
			writeError(w, http.StatusInternalServerError, "failed to create analysis")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toAnalysisResponse(analysis))
}

type ciCompareRequest struct {
	ProjectID       *string `json:"project_id,omitempty"`
	FromAnalysisID  string  `json:"from_analysis_id"`
	ToAnalysisID    string  `json:"to_analysis_id"`
	IncludeMarkdown bool    `json:"include_markdown"`
	IncludeJSON     bool    `json:"include_json"`
	UIBaseURL       *string `json:"ui_base_url,omitempty"`
}

func (h *CIHandler) CompareReport(w http.ResponseWriter, r *http.Request) {
	projectID, user, ciToken, err := resolveProject(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ciCompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if projectID == nil {
		if req.ProjectID == nil || *req.ProjectID == "" {
			writeError(w, http.StatusBadRequest, "project_id is required")
			return
		}
		parsed, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		projectID = &parsed
	} else if req.ProjectID != nil && *req.ProjectID != "" && *req.ProjectID != projectID.String() {
		writeError(w, http.StatusForbidden, "token does not match project")
		return
	}

	fromID, err := uuid.Parse(req.FromAnalysisID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from_analysis_id")
		return
	}
	toID, err := uuid.Parse(req.ToAnalysisID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to_analysis_id")
		return
	}

	// For session users, ensure access by performing standard comparison first.
	if ciToken == nil {
		if _, err := h.analyses.CompareAnalyses(r.Context(), user.ID, *projectID, fromID, toID); err != nil {
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
	}

	report, err := h.ciService.Compare(r.Context(), *projectID, ci.CompareInput{
		FromAnalysisID:  fromID,
		ToAnalysisID:    toID,
		IncludeMarkdown: req.IncludeMarkdown,
		IncludeJSON:     req.IncludeJSON,
		UIBaseURL:       req.UIBaseURL,
	})
	if err != nil {
		switch {
		case errors.Is(err, analyses.ErrAnalysisNotFound):
			writeError(w, http.StatusNotFound, "analysis not found")
		case errors.Is(err, analyses.ErrAnalysesDifferentImage):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, analyses.ErrAnalysesNotCompleted):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to build compare report")
		}
		return
	}

	writeJSON(w, http.StatusOK, report)
}

type ciCommentRequest struct {
	ProjectID      *string `json:"project_id,omitempty"`
	Provider       string  `json:"provider"`
	Repo           string  `json:"repo"`
	PRNumber       *int    `json:"pr_number,omitempty"`
	MRIID          *int    `json:"mr_iid,omitempty"`
	SCMToken       string  `json:"scm_token"`
	BodyMarkdown   string  `json:"body_markdown,omitempty"`
	ReportMarkdown *string `json:"report_markdown,omitempty"`
	ToAnalysisID   string  `json:"to_analysis_id"`
}

func (h *CIHandler) resolveRegistry(ctx context.Context, projectID uuid.UUID, registryID string, registryName *string, registryHost *string) (uuid.UUID, error) {
	var idPtr *uuid.UUID
	if registryID != "" {
		parsed, err := uuid.Parse(registryID)
		if err != nil {
			return uuid.Nil, errors.New("invalid registry id")
		}
		idPtr = &parsed
	}
	reg, err := h.registries.ResolveRegistryReference(ctx, projectID, idPtr, registryName, registryHost)
	if err != nil {
		switch {
		case errors.Is(err, registries.ErrRegistryNotFound):
			return uuid.Nil, errors.New("registry not found")
		case errors.Is(err, registries.ErrRegistryAmbiguous):
			return uuid.Nil, errors.New("multiple registries match name")
		case errors.Is(err, registries.ErrMissingRegistry):
			return uuid.Nil, errors.New("missing registry identifier")
		}
		return uuid.Nil, err
	}
	return reg.ID, nil
}

func (h *CIHandler) PostComment(w http.ResponseWriter, r *http.Request) {
	projectID, user, ciToken, err := resolveProject(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ciCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if projectID == nil {
		if req.ProjectID == nil || *req.ProjectID == "" {
			writeError(w, http.StatusBadRequest, "project_id is required")
			return
		}
		parsed, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		projectID = &parsed
	} else if req.ProjectID != nil && *req.ProjectID != "" && *req.ProjectID != projectID.String() {
		writeError(w, http.StatusForbidden, "token does not match project")
		return
	}

	if req.SCMToken == "" {
		writeError(w, http.StatusBadRequest, "scm_token is required")
		return
	}
	if req.Provider == "" || req.Repo == "" {
		writeError(w, http.StatusBadRequest, "provider and repo are required")
		return
	}
	body := strings.TrimSpace(req.BodyMarkdown)
	if body == "" && req.ReportMarkdown != nil {
		body = strings.TrimSpace(*req.ReportMarkdown)
	}
	if body == "" {
		writeError(w, http.StatusBadRequest, "body_markdown is required")
		return
	}
	toID, err := uuid.Parse(req.ToAnalysisID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to_analysis_id")
		return
	}

	if ciToken == nil {
		// ensure user has access to the analysis
		if _, err := h.analyses.GetAnalysis(r.Context(), user.ID, *projectID, toID); err != nil {
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
	} else {
		if _, err := h.analyses.GetAnalysisForProject(r.Context(), *projectID, toID); err != nil {
			if errors.Is(err, analyses.ErrAnalysisNotFound) {
				writeError(w, http.StatusNotFound, "analysis not found")
			} else {
				writeError(w, http.StatusInternalServerError, "failed to fetch analysis")
			}
			return
		}
	}

	if err := h.ciService.PostComment(r.Context(), ci.CommentInput{
		Provider:     req.Provider,
		Repo:         req.Repo,
		PRNumber:     req.PRNumber,
		MRIID:        req.MRIID,
		SCMToken:     req.SCMToken,
		BodyMarkdown: body,
		ProjectID:    *projectID,
		ToAnalysisID: toID,
	}); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// resolveProject determines project ID based on CI token or user payload.
func resolveProject(r *http.Request) (*uuid.UUID, *auth.User, *citokens.Token, error) {
	if token, ok := citokens.TokenFromContext(r.Context()); ok {
		id := token.ProjectID
		return &id, nil, &token, nil
	}

	if user, ok := auth.UserFromContext(r.Context()); ok {
		return nil, &user, nil, nil
	}

	return nil, nil, nil, errors.New("unauthorized")
}
