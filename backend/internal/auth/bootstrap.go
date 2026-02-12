package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type BootstrapAdminConfig struct {
	Email        string
	Username     string
	Password     string
	PasswordHash string
}

type BootstrapUserStore interface {
	HasAnyAdmin(ctx context.Context) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUser(ctx context.Context, login, email, passwordHash string) (User, error)
	SetUserAdmin(ctx context.Context, id string, isAdmin bool) error
}

// EnsureBootstrapAdmin provisions the first admin user when no admin exists.
// It is idempotent and does nothing if at least one admin is already present.
func EnsureBootstrapAdmin(ctx context.Context, users BootstrapUserStore, cfg BootstrapAdminConfig) (bool, error) {
	hasAdmin, err := users.HasAnyAdmin(ctx)
	if err != nil {
		return false, err
	}
	if hasAdmin {
		return false, nil
	}

	email := normalizeEmail(cfg.Email)
	if email == "" {
		return false, nil
	}
	if !isValidEmail(email) {
		return false, errors.New("DOCKSLIM_BOOTSTRAP_ADMIN_EMAIL must be a valid email")
	}

	passwordHash, err := resolveBootstrapPasswordHash(cfg.Password, cfg.PasswordHash)
	if err != nil {
		return false, err
	}

	existing, err := users.GetUserByEmail(ctx, email)
	if err == nil {
		if existing.IsAdmin {
			return false, nil
		}
		if err := users.SetUserAdmin(ctx, existing.ID.String(), true); err != nil {
			return false, err
		}
		return true, nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return false, err
	}

	baseLogin, explicitUsername, err := resolveBootstrapLogin(cfg.Username, email)
	if err != nil {
		return false, err
	}
	const maxAttempts = 100
	attempts := maxAttempts
	if explicitUsername {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		login := loginAttempt(baseLogin, i)
		created, err := users.CreateUser(ctx, login, email, passwordHash)
		if err == nil {
			if err := users.SetUserAdmin(ctx, created.ID.String(), true); err != nil {
				return false, err
			}
			return true, nil
		}

		switch {
		case errors.Is(err, ErrLoginAlreadyExists):
			if explicitUsername {
				return false, errors.New("DOCKSLIM_BOOTSTRAP_ADMIN_USERNAME already exists")
			}
			continue
		case errors.Is(err, ErrEmailAlreadyExists):
			found, getErr := users.GetUserByEmail(ctx, email)
			if getErr != nil {
				return false, getErr
			}
			if found.IsAdmin {
				return false, nil
			}
			if setErr := users.SetUserAdmin(ctx, found.ID.String(), true); setErr != nil {
				return false, setErr
			}
			return true, nil
		default:
			return false, err
		}
	}

	return false, fmt.Errorf("failed to create bootstrap admin user after %d attempts", maxAttempts)
}

func resolveBootstrapPasswordHash(password, passwordHash string) (string, error) {
	hash := strings.TrimSpace(passwordHash)
	if hash != "" {
		if _, err := bcrypt.Cost([]byte(hash)); err != nil {
			return "", errors.New("DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD_HASH must be a valid bcrypt hash")
		}
		return hash, nil
	}

	plain := strings.TrimSpace(password)
	if plain == "" {
		return "", errors.New("bootstrap admin requires DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD or DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD_HASH")
	}
	if len(plain) < 8 {
		return "", errors.New("DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD must be at least 8 characters")
	}
	return HashPassword(plain)
}

var bootstrapLoginSanitizer = regexp.MustCompile(`[^a-z0-9._-]+`)

func resolveBootstrapLogin(username, email string) (string, bool, error) {
	cleanUsername := normalizeLogin(username)
	if cleanUsername != "" {
		if !isValidLogin(cleanUsername) {
			return "", false, errors.New("DOCKSLIM_BOOTSTRAP_ADMIN_USERNAME must match ^[a-z0-9._-]{3,32}$")
		}
		return cleanUsername, true, nil
	}
	return bootstrapLoginFromEmail(email), false, nil
}

func bootstrapLoginFromEmail(email string) string {
	local := email
	if at := strings.Index(local, "@"); at > 0 {
		local = local[:at]
	}
	local = strings.ToLower(strings.TrimSpace(local))
	local = bootstrapLoginSanitizer.ReplaceAllString(local, "-")
	local = strings.Trim(local, "._-")
	if local == "" {
		local = "admin"
	}
	if len(local) > 32 {
		local = local[:32]
	}
	for len(local) < 3 {
		local += "a"
	}
	return local
}

func loginAttempt(base string, attempt int) string {
	if attempt == 0 {
		return base
	}
	suffix := "-" + strconv.Itoa(attempt)
	maxBaseLen := 32 - len(suffix)
	if maxBaseLen < 1 {
		maxBaseLen = 1
	}
	trimmed := base
	if len(trimmed) > maxBaseLen {
		trimmed = trimmed[:maxBaseLen]
	}
	trimmed = strings.Trim(trimmed, "._-")
	if trimmed == "" {
		trimmed = "admin"
	}
	login := trimmed + suffix
	if len(login) > 32 {
		login = login[:32]
	}
	for len(login) < 3 {
		login += "a"
	}
	return login
}
