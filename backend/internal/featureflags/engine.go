package featureflags

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Engine struct {
	store store
	nowFn func() time.Time
}

func NewEngine(store store) *Engine {
	return &Engine{
		store: store,
		nowFn: time.Now,
	}
}

func (e *Engine) GetUserFeatures(ctx context.Context, userID uuid.UUID) (UserFeatures, error) {
	isAdmin, err := e.store.IsAdmin(ctx, userID)
	if err != nil {
		return UserFeatures{}, err
	}

	sub, err := e.store.GetUserSubscriptionWithPlan(ctx, userID)
	switch {
	case err == nil:
	case errors.Is(err, ErrSubscriptionNotFound):
		sub, err = e.loadDefaultPlan(ctx, userID)
		if err != nil {
			return UserFeatures{}, err
		}
	default:
		return UserFeatures{}, err
	}

	if sub.Status != SubscriptionStatusActive || (sub.ValidUntil != nil && sub.ValidUntil.Before(e.nowFn())) {
		sub, err = e.loadDefaultPlan(ctx, userID)
		if err != nil {
			return UserFeatures{}, err
		}
	}

	features := cloneFeatures(sub.Features)
	if isAdmin {
		plans, err := e.store.ListPlans(ctx)
		if err != nil {
			return UserFeatures{}, err
		}
		features = mergeMaxAccessFeatures(plans)
		if planID, planName := selectAdminPlan(plans); planID != "" {
			sub.PlanID = planID
			sub.PlanName = planName
			sub.Status = SubscriptionStatusActive
			sub.ValidUntil = nil
		}
	}

	return UserFeatures{
		UserID:     userID,
		PlanID:     sub.PlanID,
		PlanName:   sub.PlanName,
		Status:     sub.Status,
		ValidUntil: sub.ValidUntil,
		IsAdmin:    isAdmin,
		Features:   features,
	}, nil
}

func (e *Engine) HasFeature(ctx context.Context, userID uuid.UUID, featureName string) (bool, error) {
	featureSet, err := e.GetUserFeatures(ctx, userID)
	if err != nil {
		return false, err
	}
	value, ok := featureSet.FeatureValue(featureName)
	if !ok {
		return false, nil
	}
	return FeatureEnabled(value), nil
}

func (e *Engine) UpdateUserSubscription(ctx context.Context, input UpdateSubscriptionInput) (UserFeatures, error) {
	if input.UserID == uuid.Nil {
		return UserFeatures{}, fmt.Errorf("user_id is required")
	}

	input.PlanID = strings.TrimSpace(strings.ToLower(input.PlanID))
	if input.PlanID == "" {
		return UserFeatures{}, fmt.Errorf("plan_id is required")
	}
	if input.Status == "" {
		input.Status = SubscriptionStatusActive
	}
	input.Status = strings.TrimSpace(strings.ToLower(input.Status))
	switch input.Status {
	case SubscriptionStatusActive, SubscriptionStatusExpired:
	default:
		return UserFeatures{}, fmt.Errorf("invalid status")
	}

	if _, err := e.store.GetPlanByID(ctx, input.PlanID); err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			return UserFeatures{}, err
		}
		return UserFeatures{}, fmt.Errorf("failed to validate plan: %w", err)
	}

	if _, err := e.store.UpsertUserSubscription(ctx, input); err != nil {
		return UserFeatures{}, err
	}

	return e.GetUserFeatures(ctx, input.UserID)
}

func FeatureEnabled(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		normalized := strings.TrimSpace(strings.ToLower(v))
		return normalized != "" && normalized != "false" && normalized != "off" && normalized != "disabled" && normalized != "none" && normalized != "0"
	case float64:
		return v > 0
	case float32:
		return v > 0
	case int:
		return v > 0
	case int8:
		return v > 0
	case int16:
		return v > 0
	case int32:
		return v > 0
	case int64:
		return v > 0
	case uint:
		return v > 0
	case uint8:
		return v > 0
	case uint16:
		return v > 0
	case uint32:
		return v > 0
	case uint64:
		return v > 0
	case nil:
		return false
	default:
		return false
	}
}

func (e *Engine) loadDefaultPlan(ctx context.Context, userID uuid.UUID) (subscriptionWithPlan, error) {
	plan, err := e.store.GetPlanByID(ctx, PlanFree)
	if err != nil {
		return subscriptionWithPlan{}, err
	}
	return subscriptionWithPlan{
		UserSubscription: UserSubscription{
			UserID: userID,
			PlanID: plan.ID,
			Status: SubscriptionStatusActive,
		},
		PlanName: plan.Name,
		Features: cloneFeatures(plan.Features),
	}, nil
}

func cloneFeatures(features map[string]any) map[string]any {
	if len(features) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(features))
	for key, value := range features {
		cloned[key] = value
	}
	return cloned
}

func mergeMaxAccessFeatures(plans []Plan) map[string]any {
	merged := make(map[string]any)
	for _, plan := range plans {
		for key, value := range plan.Features {
			existing, ok := merged[key]
			if !ok {
				merged[key] = value
				continue
			}
			merged[key] = pickMorePermissive(existing, value)
		}
	}
	return merged
}

func selectAdminPlan(plans []Plan) (string, string) {
	byID := make(map[string]Plan, len(plans))
	for _, plan := range plans {
		byID[strings.ToLower(plan.ID)] = plan
	}
	for _, candidate := range []string{PlanTeam, PlanPro, PlanFree} {
		if plan, ok := byID[candidate]; ok {
			return plan.ID, plan.Name
		}
	}
	return "", ""
}

func pickMorePermissive(current, candidate any) any {
	if current == nil || candidate == nil {
		return nil
	}

	if FeatureEnabled(current) && !FeatureEnabled(candidate) {
		return current
	}
	if !FeatureEnabled(current) && FeatureEnabled(candidate) {
		return candidate
	}

	currentNumber, currentIsNumber := asFloat(current)
	candidateNumber, candidateIsNumber := asFloat(candidate)
	if currentIsNumber && candidateIsNumber {
		if candidateNumber > currentNumber {
			return candidate
		}
		return current
	}

	currentText := strings.TrimSpace(strings.ToLower(fmt.Sprint(current)))
	candidateText := strings.TrimSpace(strings.ToLower(fmt.Sprint(candidate)))
	if currentText == CICommentsModeLimited && candidateText == "true" {
		return true
	}
	if currentText == "true" && candidateText == CICommentsModeLimited {
		return true
	}

	return candidate
}

func asFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
