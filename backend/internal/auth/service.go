package auth

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidLogin       = errors.New("invalid login")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrLoginAlreadyExists = errors.New("login already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserStore interface {
	CreateUser(ctx context.Context, login, email, passwordHash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByLogin(ctx context.Context, login string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
}

type TokenIssuer interface {
	GenerateAccessToken(ctx context.Context, user User) (string, error)
}

type Service struct {
	users  UserStore
	tokens TokenIssuer
}

func NewService(users UserStore, tokens TokenIssuer) *Service {
	return &Service{users: users, tokens: tokens}
}

var loginRegex = regexp.MustCompile(`^[a-z0-9._-]{3,32}$`)

func (s *Service) Register(ctx context.Context, login, email, password string) (User, error) {
	cleanLogin := normalizeLogin(login)
	if !isValidLogin(cleanLogin) {
		return User{}, ErrInvalidLogin
	}
	cleanEmail := normalizeEmail(email)
	if !isValidEmail(cleanEmail) {
		return User{}, ErrInvalidEmail
	}
	if len(password) < 8 {
		return User{}, ErrInvalidPassword
	}

	if _, err := s.users.GetUserByLogin(ctx, cleanLogin); err == nil {
		return User{}, ErrLoginAlreadyExists
	} else if err != nil && !errors.Is(err, ErrUserNotFound) {
		return User{}, err
	}

	if _, err := s.users.GetUserByEmail(ctx, cleanEmail); err == nil {
		return User{}, ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, ErrUserNotFound) {
		return User{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	return s.users.CreateUser(ctx, cleanLogin, cleanEmail, hash)
}

func (s *Service) Login(ctx context.Context, identifier, password string) (User, string, error) {
	if strings.TrimSpace(identifier) == "" {
		return User{}, "", ErrInvalidCredentials
	}

	var (
		user User
		err  error
	)

	cleanEmail := normalizeEmail(identifier)
	if isValidEmail(cleanEmail) {
		user, err = s.users.GetUserByEmail(ctx, cleanEmail)
	} else {
		cleanLogin := normalizeLogin(identifier)
		if cleanLogin == "" {
			return User{}, "", ErrInvalidCredentials
		}
		user, err = s.users.GetUserByLogin(ctx, cleanLogin)
	}

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return User{}, "", ErrInvalidCredentials
		}
		return User{}, "", err
	}

	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return User{}, "", ErrInvalidCredentials
	}

	token, err := s.tokens.GenerateAccessToken(ctx, user)
	if err != nil {
		return User{}, "", err
	}

	return user, token, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeLogin(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func isValidLogin(login string) bool {
	return loginRegex.MatchString(login)
}
