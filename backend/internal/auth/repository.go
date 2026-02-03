package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAuthKeyNotFound = errors.New("auth key not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, login, email, passwordHash string) (User, error) {
	const query = `
		INSERT INTO users (login, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, login, email, password_hash, created_at, updated_at
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, login, email, passwordHash).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return User{}, ErrEmailAlreadyExists
			case "idx_users_login_unique", "users_login_key":
				return User{}, ErrLoginAlreadyExists
			}
		}
		return User{}, err
	}

	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, login, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (User, error) {
	const query = `
		SELECT id, login, email, password_hash, created_at, updated_at
		FROM users
		WHERE login = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (User, error) {
	const query = `
		SELECT id, login, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, id, login, email string) (User, error) {
	const query = `
		UPDATE users
		SET login = $2, email = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, login, email, password_hash, created_at, updated_at
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, id, login, email).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return User{}, ErrEmailAlreadyExists
			case "idx_users_login_unique", "users_login_key":
				return User{}, ErrLoginAlreadyExists
			}
		}
		return User{}, err
	}

	return user, nil
}

func (r *Repository) ListActiveKeys(ctx context.Context) ([]AuthKey, error) {
	const query = `
		SELECT id, key_id, signing_key, algorithm, is_active, created_at, rotated_at
		FROM auth_keys
		WHERE is_active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []AuthKey
	for rows.Next() {
		var key AuthKey
		if err := rows.Scan(
			&key.ID,
			&key.KeyID,
			&key.SigningKey,
			&key.Algorithm,
			&key.IsActive,
			&key.CreatedAt,
			&key.RotatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *Repository) GetKeyByID(ctx context.Context, keyID string) (AuthKey, error) {
	const query = `
		SELECT id, key_id, signing_key, algorithm, is_active, created_at, rotated_at
		FROM auth_keys
		WHERE key_id = $1
	`

	var key AuthKey
	err := r.db.QueryRowContext(ctx, query, keyID).Scan(
		&key.ID,
		&key.KeyID,
		&key.SigningKey,
		&key.Algorithm,
		&key.IsActive,
		&key.CreatedAt,
		&key.RotatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return AuthKey{}, ErrAuthKeyNotFound
	}
	if err != nil {
		return AuthKey{}, err
	}

	return key, nil
}

func (r *Repository) CreateAuthKey(ctx context.Context, key AuthKey) (AuthKey, error) {
	const query = `
		INSERT INTO auth_keys (key_id, signing_key, algorithm, is_active, created_at, rotated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		key.KeyID,
		key.SigningKey,
		key.Algorithm,
		key.IsActive,
		key.CreatedAt,
		key.RotatedAt,
	).Scan(&key.ID, &key.CreatedAt)
	if err != nil {
		return AuthKey{}, err
	}

	return key, nil
}
