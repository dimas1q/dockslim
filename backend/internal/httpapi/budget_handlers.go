package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/budgets"
)

type BudgetsHandler struct {
	service *budgets.Service
}

func NewBudgetsHandler(service *budgets.Service) *BudgetsHandler {
	return &BudgetsHandler{service: service}
}

type thresholdsRequest struct {
	WarnDeltaMB *int64 `json:"warn_delta_mb"`
	FailDeltaMB *int64 `json:"fail_delta_mb"`
	HardLimitMB *int64 `json:"hard_limit_mb"`
}

type overrideRequest struct {
	Image string `json:"image"`
	thresholdsRequest
}

func (h *BudgetsHandler) List(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.service.GetBudgets(r.Context(), user.ID, projectID)
	if err != nil {
		if errors.Is(err, budgets.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch budgets")
		return
	}

	var defaultBudget *budgetResponse
	overrides := make([]budgetResponse, 0)
	for _, item := range items {
		resp := toBudgetResponse(item)
		if item.Image == nil {
			defaultBudget = &resp
		} else {
			overrides = append(overrides, resp)
		}
	}

	payload := map[string]interface{}{
		"default":   defaultBudget,
		"overrides": overrides,
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h *BudgetsHandler) UpsertDefault(w http.ResponseWriter, r *http.Request) {
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

	var req thresholdsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	thresholds, err := thresholdsFromMB(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	budget, err := h.service.UpsertDefault(r.Context(), user.ID, projectID, budgets.DefaultBudgetInput{Thresholds: thresholds})
	if err != nil {
		switch {
		case errors.Is(err, budgets.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, budgets.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, budgets.ErrInvalidThreshold):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to save budget")
		}
		return
	}

	writeJSON(w, http.StatusOK, toBudgetResponse(budget))
}

func (h *BudgetsHandler) CreateOverride(w http.ResponseWriter, r *http.Request) {
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

	var req overrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	thresholds, err := thresholdsFromMB(req.thresholdsRequest)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	budget, err := h.service.CreateOverride(r.Context(), user.ID, projectID, budgets.OverrideBudgetInput{
		Image:      req.Image,
		Thresholds: thresholds,
	})
	if err != nil {
		switch {
		case errors.Is(err, budgets.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, budgets.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, budgets.ErrBudgetConflict), isUniqueViolation(err):
			writeError(w, http.StatusConflict, "budget override for this image already exists")
		case errors.Is(err, budgets.ErrInvalidImage), errors.Is(err, budgets.ErrInvalidThreshold):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create budget override")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toBudgetResponse(budget))
}

func (h *BudgetsHandler) UpdateOverride(w http.ResponseWriter, r *http.Request) {
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
	budgetID, err := parseUUIDParam(r, "budgetId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	var req overrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	thresholds, err := thresholdsFromMB(req.thresholdsRequest)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var image *string
	if req.Image != "" {
		image = &req.Image
	}

	budget, err := h.service.UpdateBudget(r.Context(), user.ID, projectID, budgetID, budgets.UpdateBudgetInput{
		Image:      image,
		Thresholds: thresholds,
	})
	if err != nil {
		switch {
		case errors.Is(err, budgets.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, budgets.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, budgets.ErrInvalidImage), errors.Is(err, budgets.ErrInvalidThreshold), errors.Is(err, budgets.ErrInvalidBudgetPatch):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, budgets.ErrBudgetNotFound):
			writeError(w, http.StatusNotFound, "budget not found")
		case isUniqueViolation(err):
			writeError(w, http.StatusConflict, "budget override for this image already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update budget")
		}
		return
	}

	writeJSON(w, http.StatusOK, toBudgetResponse(budget))
}

func (h *BudgetsHandler) DeleteOverride(w http.ResponseWriter, r *http.Request) {
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
	budgetID, err := parseUUIDParam(r, "budgetId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	if err := h.service.DeleteBudget(r.Context(), user.ID, projectID, budgetID); err != nil {
		switch {
		case errors.Is(err, budgets.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, budgets.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, budgets.ErrBudgetNotFound):
			writeError(w, http.StatusNotFound, "budget not found")
		case isUniqueViolation(err):
			writeError(w, http.StatusConflict, "budget override for this image already exists")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete budget")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func thresholdsFromMB(req thresholdsRequest) (budgets.ThresholdsInput, error) {
	warnBytes, err := budgets.MBToBytes(req.WarnDeltaMB)
	if err != nil {
		return budgets.ThresholdsInput{}, err
	}
	failBytes, err := budgets.MBToBytes(req.FailDeltaMB)
	if err != nil {
		return budgets.ThresholdsInput{}, err
	}
	hardBytes, err := budgets.MBToBytes(req.HardLimitMB)
	if err != nil {
		return budgets.ThresholdsInput{}, err
	}
	return budgets.ThresholdsInput{
		WarnDeltaBytes: warnBytes,
		FailDeltaBytes: failBytes,
		HardLimitBytes: hardBytes,
	}, nil
}

type budgetResponse struct {
	ID             string  `json:"id"`
	Image          *string `json:"image,omitempty"`
	WarnDeltaBytes *int64  `json:"warn_delta_bytes,omitempty"`
	FailDeltaBytes *int64  `json:"fail_delta_bytes,omitempty"`
	HardLimitBytes *int64  `json:"hard_limit_bytes,omitempty"`
	WarnDeltaMB    *int64  `json:"warn_delta_mb,omitempty"`
	FailDeltaMB    *int64  `json:"fail_delta_mb,omitempty"`
	HardLimitMB    *int64  `json:"hard_limit_mb,omitempty"`
}

func toBudgetResponse(b budgets.Budget) budgetResponse {
	conv := func(v *int64) *int64 {
		if v == nil {
			return nil
		}
		mb := *v / (1024 * 1024)
		return &mb
	}

	resp := budgetResponse{ID: b.ID.String(), Image: b.Image}
	resp.WarnDeltaBytes = b.WarnDeltaBytes
	resp.FailDeltaBytes = b.FailDeltaBytes
	resp.HardLimitBytes = b.HardLimitBytes
	resp.WarnDeltaMB = conv(b.WarnDeltaBytes)
	resp.FailDeltaMB = conv(b.FailDeltaBytes)
	resp.HardLimitMB = conv(b.HardLimitBytes)
	return resp
}
