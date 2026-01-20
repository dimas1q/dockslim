package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/dimas1q/dockslim/analyzer/internal/analysis"
	"github.com/dimas1q/dockslim/analyzer/internal/registry"
	"github.com/google/uuid"
)

const (
	jobStatusQueued  = "queued"
	jobStatusRunning = "running"
	jobStatusDone    = "done"
	jobStatusFailed  = "failed"

	analysisStatusRunning   = "running"
	analysisStatusCompleted = "completed"
	analysisStatusFailed    = "failed"
)

type Worker struct {
	db           *sql.DB
	client       *registry.Client
	workerID     string
	lockTimeout  time.Duration
	pollInterval time.Duration
}

type Job struct {
	ID         uuid.UUID
	AnalysisID uuid.UUID
}

type AnalysisInput struct {
	ID          uuid.UUID
	Image       string
	Tag         string
	RegistryURL string
	Username    string
	PasswordEnc []byte
}

func New(db *sql.DB) *Worker {
	return &Worker{
		db:           db,
		client:       registry.NewClient(),
		workerID:     uuid.NewString(),
		lockTimeout:  10 * time.Minute,
		pollInterval: 2 * time.Second,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		job, err := w.claimJob(ctx)
		if err != nil {
			return err
		}
		if job == nil {
			time.Sleep(w.pollInterval)
			continue
		}

		if err := w.processJob(ctx, *job); err != nil {
			log.Printf("analysis job %s failed: %v", job.ID, err)
		}
	}
}

func (w *Worker) claimJob(ctx context.Context) (*Job, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	staleBefore := time.Now().Add(-w.lockTimeout)
	const selectQuery = `
		SELECT id, analysis_id
		FROM analysis_jobs
		WHERE (
			status = $1
			OR (status = $2 AND (locked_at IS NULL OR locked_at < $3))
		)
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	var job Job
	err = tx.QueryRowContext(ctx, selectQuery, jobStatusQueued, jobStatusRunning, staleBefore).Scan(&job.ID, &job.AnalysisID)
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	const updateJobQuery = `
		UPDATE analysis_jobs
		SET status = $1,
			attempts = attempts + 1,
			locked_by = $2,
			locked_at = NOW(),
			updated_at = NOW()
		WHERE id = $3
	`
	if _, err = tx.ExecContext(ctx, updateJobQuery, jobStatusRunning, w.workerID, job.ID); err != nil {
		return nil, err
	}

	const updateAnalysisQuery = `
		UPDATE image_analyses
		SET status = $1,
			started_at = COALESCE(started_at, NOW()),
			updated_at = NOW()
		WHERE id = $2
	`
	if _, err = tx.ExecContext(ctx, updateAnalysisQuery, analysisStatusRunning, job.AnalysisID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &job, nil
}

func (w *Worker) processJob(ctx context.Context, job Job) error {
	input, err := w.fetchAnalysisInput(ctx, job.AnalysisID)
	if err != nil {
		_ = w.failJob(ctx, job.ID, job.AnalysisID, err)
		return err
	}

	normalizedImage, err := normalizeImageReference(input.Image, input.RegistryURL)
	if err != nil {
		_ = w.failJob(ctx, job.ID, job.AnalysisID, err)
		return err
	}
	input.Image = normalizedImage

	password, err := w.decryptPassword(ctx, input.PasswordEnc)
	if err != nil {
		_ = w.failJob(ctx, job.ID, job.AnalysisID, err)
		return err
	}

	if err := w.client.Ping(ctx, input.RegistryURL, input.Username, password); err != nil {
		_ = w.failJob(ctx, job.ID, job.AnalysisID, err)
		return err
	}

	manifestSummary, err := w.client.FetchManifest(ctx, input.RegistryURL, input.Image, input.Tag, input.Username, password)
	if err != nil {
		_ = w.failJob(ctx, job.ID, job.AnalysisID, err)
		return err
	}

	layers := make([]analysis.LayerResult, 0, len(manifestSummary.Layers))
	for _, layer := range manifestSummary.Layers {
		layers = append(layers, analysis.LayerResult{
			Digest:    layer.Digest,
			SizeBytes: layer.Size,
			MediaType: layer.MediaType,
		})
	}

	totalSize := manifestSummary.TotalSize
	insights := analysis.BuildInsights(manifestSummary.Layers, totalSize)
	recommendations := analysis.BuildRecommendations(manifestSummary.Layers, totalSize, manifestSummary.MediaType)

	result := analysis.Result{
		Image:           input.Image,
		Tag:             input.Tag,
		MediaType:       manifestSummary.MediaType,
		Layers:          layers,
		TotalSizeBytes:  totalSize,
		Insights:        insights,
		Recommendations: recommendations,
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		_ = w.failJob(ctx, job.ID, job.AnalysisID, err)
		return err
	}

	if err := w.completeJob(ctx, job.ID, job.AnalysisID, resultJSON, &totalSize); err != nil {
		return err
	}

	return nil
}

func (w *Worker) fetchAnalysisInput(ctx context.Context, analysisID uuid.UUID) (AnalysisInput, error) {
	const query = `
		SELECT ia.id, ia.image, ia.tag, r.registry_url, r.username, r.password_enc
		FROM image_analyses ia
		JOIN registries r ON ia.registry_id = r.id
		WHERE ia.id = $1
	`

	var input AnalysisInput
	var username sql.NullString
	var passwordEnc []byte
	err := w.db.QueryRowContext(ctx, query, analysisID).Scan(
		&input.ID,
		&input.Image,
		&input.Tag,
		&input.RegistryURL,
		&username,
		&passwordEnc,
	)
	if err != nil {
		return AnalysisInput{}, err
	}

	if username.Valid {
		input.Username = username.String
	}
	if len(passwordEnc) > 0 {
		input.PasswordEnc = passwordEnc
	}

	return input, nil
}

func (w *Worker) decryptPassword(ctx context.Context, encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	const query = `
		SELECT key_material
		FROM encryption_keys
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`
	var keyMaterial []byte
	if err := w.db.QueryRowContext(ctx, query).Scan(&keyMaterial); err != nil {
		return "", err
	}

	return registry.DecryptSecret(keyMaterial, encrypted)
}

func (w *Worker) completeJob(ctx context.Context, jobID, analysisID uuid.UUID, resultJSON []byte, totalSize *int64) error {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var size sql.NullInt64
	if totalSize != nil {
		size = sql.NullInt64{Int64: *totalSize, Valid: true}
	}

	const updateAnalysisQuery = `
		UPDATE image_analyses
		SET status = $1,
			total_size_bytes = $2,
			result_json = $3,
			finished_at = NOW(),
			updated_at = NOW()
		WHERE id = $4
	`
	if _, err = tx.ExecContext(ctx, updateAnalysisQuery, analysisStatusCompleted, size, resultJSON, analysisID); err != nil {
		return err
	}

	const updateJobQuery = `
		UPDATE analysis_jobs
		SET status = $1,
			last_error = NULL,
			updated_at = NOW()
		WHERE id = $2
	`
	if _, err = tx.ExecContext(ctx, updateJobQuery, jobStatusDone, jobID); err != nil {
		return err
	}

	return tx.Commit()
}

func normalizeImageReference(image, registryURL string) (string, error) {
	if strings.Contains(image, "://") {
		return "", fmt.Errorf("invalid image reference")
	}

	parts := strings.SplitN(image, "/", 2)
	if len(parts) < 2 {
		return image, nil
	}

	hostPart := parts[0]
	if !looksLikeRegistryHost(hostPart) {
		return image, nil
	}

	registryHost, err := extractRegistryHostname(registryURL)
	if err != nil {
		return "", err
	}

	imageHost := extractImageHostname(hostPart)
	if !strings.EqualFold(imageHost, registryHost) {
		return "", fmt.Errorf("image registry does not match selected registry")
	}

	if strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("invalid image reference")
	}

	return parts[1], nil
}

func looksLikeRegistryHost(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, ".") || strings.Contains(lower, ":") || lower == "localhost"
}

func extractRegistryHostname(registryURL string) (string, error) {
	parsed, err := url.Parse(registryURL)
	if err != nil {
		return "", err
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("invalid registry url")
	}
	return parsed.Hostname(), nil
}

func extractImageHostname(hostPart string) string {
	if strings.HasPrefix(hostPart, "[") {
		return strings.Trim(hostPart, "[]")
	}
	if strings.Contains(hostPart, ":") {
		parts := strings.Split(hostPart, ":")
		return parts[0]
	}
	return hostPart
}

func (w *Worker) failJob(ctx context.Context, jobID, analysisID uuid.UUID, failure error) error {
	result := map[string]any{
		"error": failure.Error(),
	}
	resultJSON, _ := json.Marshal(result)

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const updateAnalysisQuery = `
		UPDATE image_analyses
		SET status = $1,
			result_json = $2,
			finished_at = NOW(),
			updated_at = NOW()
		WHERE id = $3
	`
	if _, err = tx.ExecContext(ctx, updateAnalysisQuery, analysisStatusFailed, resultJSON, analysisID); err != nil {
		return err
	}

	const updateJobQuery = `
		UPDATE analysis_jobs
		SET status = $1,
			last_error = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	if _, err = tx.ExecContext(ctx, updateJobQuery, jobStatusFailed, failure.Error(), jobID); err != nil {
		return err
	}

	return tx.Commit()
}
