package budgets

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestListBudgetsDefaultFirst(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	projectID := uuid.New()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "project_id", "image", "warn_delta_bytes", "fail_delta_bytes", "hard_limit_bytes", "created_at", "updated_at"}).
		AddRow(uuid.New(), projectID, nil, nil, nil, nil, now, now).
		AddRow(uuid.New(), projectID, "app", nil, nil, nil, now, now)

	mock.ExpectQuery("SELECT id, project_id, image, warn_delta_bytes").WithArgs(projectID).WillReturnRows(rows)

	budgets, err := repo.ListBudgetsByProject(context.Background(), projectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(budgets) != 2 {
		t.Fatalf("expected 2 budgets, got %d", len(budgets))
	}
	if budgets[0].Image != nil {
		t.Fatalf("expected default budget first")
	}
	if budgets[1].Image == nil || *budgets[1].Image != "app" {
		t.Fatalf("expected override second")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
