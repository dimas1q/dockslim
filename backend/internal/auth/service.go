package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserStore interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
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

func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	cleanEmail := normalizeEmail(email)
	if !isValidEmail(cleanEmail) {
		return User{}, ErrInvalidEmail
	}
	if len(password) < 8 {
		return User{}, ErrInvalidPassword
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

	return s.users.CreateUser(ctx, cleanEmail, hash)
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	cleanEmail := normalizeEmail(email)
	user, err := s.users.GetUserByEmail(ctx, cleanEmail)
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

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
