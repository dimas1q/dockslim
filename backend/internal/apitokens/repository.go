package apitokens

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
)

var (
	ErrTokenNotFound = errors.New("api token not found")
	ErrTokenConflict = errors.New("api token name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CreateTokenParams struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	TokenHash string
	ExpiresAt *time.Time
}

func (r *Repository) CreateToken(ctx context.Context, params CreateTokenParams) (Token, error) {
	const query = `
		INSERT INTO user_api_tokens (id, user_id, name, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	var expires interface{}
	if params.ExpiresAt != nil {
		expires = *params.ExpiresAt
	}

	token := Token{
		ID:        params.ID,
		UserID:    params.UserID,
		Name:      params.Name,
		TokenHash: params.TokenHash,
		ExpiresAt: params.ExpiresAt,
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		params.ID,
		params.UserID,
		params.Name,
		params.TokenHash,
		expires,
	).Scan(&token.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Token{}, ErrTokenConflict
		}
		return Token{}, err
	}

	return token, nil
}

func (r *Repository) ListTokensByUser(ctx context.Context, userID uuid.UUID) ([]Token, error) {
	const query = `
		SELECT id, user_id, name, token_hash, last_used_at, created_at, revoked_at, expires_at
		FROM user_api_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		token, err := scanToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *Repository) GetTokenByID(ctx context.Context, tokenID uuid.UUID) (Token, error) {
	const query = `
		SELECT id, user_id, name, token_hash, last_used_at, created_at, revoked_at, expires_at
		FROM user_api_tokens
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, tokenID)
	token, err := scanToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Token{}, ErrTokenNotFound
		}
		return Token{}, err
	}
	return token, nil
}

func (r *Repository) RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	const query = `
		UPDATE user_api_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, tokenID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

func (r *Repository) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID, ts time.Time) error {
	const query = `
		UPDATE user_api_tokens
		SET last_used_at = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, tokenID, ts)
	return err
}

type tokenRow interface {
	Scan(dest ...any) error
}

func scanToken(row tokenRow) (Token, error) {
	var token Token
	var lastUsed sql.NullTime
	var revokedAt sql.NullTime
	var expiresAt sql.NullTime

	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.Name,
		&token.TokenHash,
		&lastUsed,
		&token.CreatedAt,
		&revokedAt,
		&expiresAt,
	)
	if err != nil {
		return Token{}, err
	}

	if lastUsed.Valid {
		token.LastUsedAt = &lastUsed.Time
	}
	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}
	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Time
	}

	return token, nil
}
