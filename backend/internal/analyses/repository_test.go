package analyses

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCreateAnalysisEnqueuesJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	projectID := uuid.New()
	registryID := uuid.New()
	analysisID := uuid.New()
	now := time.Now()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{
		"id",
		"project_id",
		"registry_id",
		"image",
		"tag",
		"status",
		"total_size_bytes",
		"result_json",
		"started_at",
		"finished_at",
		"created_at",
		"updated_at",
	}).AddRow(
		analysisID,
		projectID,
		registryID.String(),
		"repo/image",
		"latest",
		StatusQueued,
		nil,
		nil,
		nil,
		nil,
		now,
		now,
	)

	mock.ExpectQuery("INSERT INTO image_analyses").
		WithArgs(projectID, registryID, "repo/image", "latest", StatusQueued, sqlmock.AnyArg()).
		WillReturnRows(rows)

	mock.ExpectExec("INSERT INTO analysis_jobs").
		WithArgs(analysisID, StatusQueued).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	_, err = repo.CreateAnalysis(context.Background(), CreateAnalysisParams{
		ProjectID:  projectID,
		RegistryID: &registryID,
		Image:      "repo/image",
		Tag:        "latest",
		Status:     StatusQueued,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRerunAnalysisResetsAndEnqueuesJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	projectID := uuid.New()
	analysisID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE image_analyses").
		WithArgs(StatusQueued, analysisID, projectID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO analysis_jobs").
		WithArgs(analysisID, StatusQueued).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.RerunAnalysis(context.Background(), projectID, analysisID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetAnalysisForProjectReturnsResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	projectID := uuid.New()
	analysisID := uuid.New()
	now := time.Now()
	totalSize := int64(98765)
	resultJSON := `{"total_size_bytes":98765}`

	rows := sqlmock.NewRows([]string{
		"id",
		"project_id",
		"registry_id",
		"image",
		"tag",
		"status",
		"total_size_bytes",
		"result_json",
		"started_at",
		"finished_at",
		"created_at",
		"updated_at",
	}).AddRow(
		analysisID,
		projectID,
		nil,
		"repo/image",
		"latest",
		StatusCompleted,
		totalSize,
		resultJSON,
		now,
		now,
		now,
		now,
	)

	mock.ExpectQuery("SELECT id, project_id, registry_id, image, tag, status, total_size_bytes, result_json, started_at, finished_at, created_at, updated_at").
		WithArgs(analysisID, projectID).
		WillReturnRows(rows)

	analysis, err := repo.GetAnalysisForProject(context.Background(), projectID, analysisID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if analysis.TotalSizeBytes == nil || *analysis.TotalSizeBytes != totalSize {
		t.Fatalf("expected total size %d, got %+v", totalSize, analysis.TotalSizeBytes)
	}
	if string(analysis.ResultJSON) != resultJSON {
		t.Fatalf("expected result json %s, got %s", resultJSON, string(analysis.ResultJSON))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
