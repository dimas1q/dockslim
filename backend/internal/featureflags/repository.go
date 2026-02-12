package featureflags

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrPlanNotFound         = errors.New("plan not found")
)

type subscriptionWithPlan struct {
	UserSubscription
	PlanName string
	Features map[string]any
}

type store interface {
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	GetUserSubscriptionWithPlan(ctx context.Context, userID uuid.UUID) (subscriptionWithPlan, error)
	GetPlanByID(ctx context.Context, planID string) (Plan, error)
	ListPlans(ctx context.Context) ([]Plan, error)
	UpsertUserSubscription(ctx context.Context, input UpdateSubscriptionInput) (UserSubscription, error)
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	const query = `
		SELECT is_admin
		FROM users
		WHERE id = $1
	`
	var isAdmin bool
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&isAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return isAdmin, nil
}

func (r *Repository) GetUserSubscriptionWithPlan(ctx context.Context, userID uuid.UUID) (subscriptionWithPlan, error) {
	const query = `
		SELECT s.user_id, s.plan_id, s.status, s.valid_until, s.created_at, s.updated_at, p.name, p.features
		FROM user_subscriptions s
		INNER JOIN plans p ON p.id = s.plan_id
		WHERE s.user_id = $1
	`

	var sub subscriptionWithPlan
	var rawFeatures []byte
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&sub.UserID,
		&sub.PlanID,
		&sub.Status,
		&sub.ValidUntil,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&sub.PlanName,
		&rawFeatures,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return subscriptionWithPlan{}, ErrSubscriptionNotFound
		}
		return subscriptionWithPlan{}, err
	}

	features, err := parseFeatures(rawFeatures)
	if err != nil {
		return subscriptionWithPlan{}, err
	}
	sub.Features = features
	return sub, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, planID string) (Plan, error) {
	const query = `
		SELECT id, name, features, created_at
		FROM plans
		WHERE id = $1
	`

	var plan Plan
	var rawFeatures []byte
	if err := r.db.QueryRowContext(ctx, query, planID).Scan(&plan.ID, &plan.Name, &rawFeatures, &plan.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Plan{}, ErrPlanNotFound
		}
		return Plan{}, err
	}

	features, err := parseFeatures(rawFeatures)
	if err != nil {
		return Plan{}, err
	}
	plan.Features = features
	return plan, nil
}

func (r *Repository) ListPlans(ctx context.Context) ([]Plan, error) {
	const query = `
		SELECT id, name, features, created_at
		FROM plans
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]Plan, 0, 4)
	for rows.Next() {
		var plan Plan
		var rawFeatures []byte
		if err := rows.Scan(&plan.ID, &plan.Name, &rawFeatures, &plan.CreatedAt); err != nil {
			return nil, err
		}
		features, err := parseFeatures(rawFeatures)
		if err != nil {
			return nil, err
		}
		plan.Features = features
		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *Repository) UpsertUserSubscription(ctx context.Context, input UpdateSubscriptionInput) (UserSubscription, error) {
	const query = `
		INSERT INTO user_subscriptions (user_id, plan_id, status, valid_until)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET
			plan_id = EXCLUDED.plan_id,
			status = EXCLUDED.status,
			valid_until = EXCLUDED.valid_until,
			updated_at = NOW()
		RETURNING user_id, plan_id, status, valid_until, created_at, updated_at
	`

	var sub UserSubscription
	err := r.db.QueryRowContext(ctx, query, input.UserID, input.PlanID, input.Status, input.ValidUntil).Scan(
		&sub.UserID,
		&sub.PlanID,
		&sub.Status,
		&sub.ValidUntil,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		return UserSubscription{}, err
	}
	return sub, nil
}

func parseFeatures(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}
