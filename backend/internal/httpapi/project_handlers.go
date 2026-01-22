package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectsHandler struct {
	service *projects.Service
}

func NewProjectsHandler(service *projects.Service) *ProjectsHandler {
	return &ProjectsHandler{service: service}
}

type projectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type projectPatchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type projectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Role        string    `json:"role,omitempty"`
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.service.CreateProject(r.Context(), user.ID, req.Name, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, projects.ErrInvalidProjectName):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create project")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toProjectResponse(project))
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectsList, err := h.service.ListProjects(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}

	resp := make([]projectResponse, 0, len(projectsList))
	for _, project := range projectsList {
		resp = append(resp, toProjectResponse(project))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseProjectID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.service.GetProject(r.Context(), user.ID, projectID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch project")
		return
	}

	role, err := h.service.GetMemberRole(r.Context(), user.ID, projectID)
	if err != nil {
		if errors.Is(err, projects.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch project role")
		return
	}

	writeJSON(w, http.StatusOK, toProjectResponseWithRole(project, role))
}

func (h *ProjectsHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseProjectID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req projectPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.service.UpdateProject(r.Context(), user.ID, projectID, projects.UpdateProjectInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, projects.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, projects.ErrInvalidProjectName):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, projects.ErrInvalidProjectPatch):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, projects.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update project")
		}
		return
	}

	writeJSON(w, http.StatusOK, toProjectResponse(project))
}

func (h *ProjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseProjectID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	if err := h.service.DeleteProject(r.Context(), user.ID, projectID); err != nil {
		switch {
		case errors.Is(err, projects.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, projects.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete project")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseProjectID(r *http.Request) (uuid.UUID, error) {
	idStr := chi.URLParam(r, "id")
	return uuid.Parse(idStr)
}

func toProjectResponse(project projects.Project) projectResponse {
	return projectResponse{
		ID:          project.ID.String(),
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func toProjectResponseWithRole(project projects.Project, role string) projectResponse {
	resp := toProjectResponse(project)
	resp.Role = role
	return resp
}
