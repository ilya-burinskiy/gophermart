package services

import (
	"context"
	"fmt"

	"github.com/ilya-burinskiy/gophermart/internal/auth"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type UserAuthenticator interface {
	Call(ctx context.Context, login, password string) (string, error)
}

type AuthenticateUserService struct {
	store storage.Storage
}

func NewAuthenticateUserService(store storage.Storage) AuthenticateUserService {
	return AuthenticateUserService{store: store}
}

func (srv AuthenticateUserService) Call(ctx context.Context, login, password string) (string, error) {
	user, err := srv.store.FindUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	if !auth.ValidatePasswordHash(password, user.EncryptedPassword) {
		return "", auth.ErrInvalidCreds
	}

	jwtStr, err := auth.BuildJWTString(user)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	return jwtStr, nil
}
