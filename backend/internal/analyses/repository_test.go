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
		"git_ref",
		"commit_sha",
		"status",
		"total_size_bytes",
		"layer_count",
		"largest_layer_bytes",
		"result_json",
		"started_at",
		"finished_at",
		"analyzed_at",
		"created_at",
		"updated_at",
	}).AddRow(
		analysisID,
		projectID,
		registryID.String(),
		"repo/image",
		"latest",
		"main",
		"abc123",
		StatusQueued,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		now,
		now,
	)

	mock.ExpectQuery("INSERT INTO image_analyses").
		WithArgs(projectID, registryID, "repo/image", "latest", sqlmock.AnyArg(), sqlmock.AnyArg(), StatusQueued, sqlmock.AnyArg()).
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
		GitRef:     func() *string { v := "main"; return &v }(),
		CommitSHA:  func() *string { v := "abc123"; return &v }(),
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
		"git_ref",
		"commit_sha",
		"status",
		"total_size_bytes",
		"layer_count",
		"largest_layer_bytes",
		"result_json",
		"started_at",
		"finished_at",
		"analyzed_at",
		"created_at",
		"updated_at",
	}).AddRow(
		analysisID,
		projectID,
		nil,
		"repo/image",
		"latest",
		"main",
		"abc123",
		StatusCompleted,
		totalSize,
		int64(5),
		int64(1234),
		resultJSON,
		now,
		now,
		now,
		now,
		now,
	)

	mock.ExpectQuery("SELECT id, project_id, registry_id, image, tag, git_ref, commit_sha, status, total_size_bytes, layer_count, largest_layer_bytes, result_json, started_at, finished_at, analyzed_at, created_at, updated_at").
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

func TestGetLatestCompletedBaseline(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	projectID := uuid.New()
	analysisID := uuid.New()
	excludeID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"project_id",
		"registry_id",
		"image",
		"tag",
		"git_ref",
		"commit_sha",
		"status",
		"total_size_bytes",
		"layer_count",
		"largest_layer_bytes",
		"result_json",
		"started_at",
		"finished_at",
		"analyzed_at",
		"created_at",
		"updated_at",
	}).AddRow(
		analysisID,
		projectID,
		nil,
		"repo/image",
		"latest",
		"main",
		"abc123",
		StatusCompleted,
		int64(1000),
		int64(5),
		int64(200),
		`{"total_size_bytes":1000}`,
		now,
		now,
		now,
		now,
		now,
	)

	mock.ExpectQuery("FROM image_analyses").
		WithArgs(projectID, "repo/image", "main", StatusCompleted, excludeID).
		WillReturnRows(rows)

	analysis, err := repo.GetLatestCompletedBaseline(context.Background(), projectID, "repo/image", "main", excludeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if analysis.ID != analysisID {
		t.Fatalf("expected analysis %s, got %s", analysisID, analysis.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListTrendsSkipsNullValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	projectID := uuid.New()
	now := time.Now()
	later := now.Add(1 * time.Hour)

	rows := sqlmock.NewRows([]string{"analyzed_at", "total_size_bytes"}).
		AddRow(now, nil).
		AddRow(later, int64(12345))

	mock.ExpectQuery("SELECT analyzed_at, total_size_bytes").
		WithArgs(projectID, StatusCompleted, 1000).
		WillReturnRows(rows)

	points, err := repo.ListTrends(context.Background(), projectID, TrendMetricTotalSize, HistoryFilter{Limit: 1000})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(points))
	}
	if points[0].Value != 12345 {
		t.Fatalf("expected value 12345, got %d", points[0].Value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
