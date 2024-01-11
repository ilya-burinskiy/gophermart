package services

import (
	"context"
	"fmt"

	"github.com/ilya-burinskiy/gophermart/internal/auth"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type UserRegistrator interface {
	Call(ctx context.Context, login, password string) (string, error)
}

type RegisterUserService struct {
	store storage.Storage
}

func NewRegisterUserService(store storage.Storage) RegisterUserService {
	return RegisterUserService{store: store}
}

func (srv RegisterUserService) Call(ctx context.Context, login, password string) (string, error) {
	encryptedPassword, err := auth.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	user, err := srv.store.CreateUser(ctx, login, encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	jwtStr, err := auth.BuildJWTString(user)
	if err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	return jwtStr, nil
}
