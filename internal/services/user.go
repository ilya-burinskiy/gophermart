package services

import (
	"context"

	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type UserOrdersFetcher interface {
	Call(ctx context.Context, userID int) ([]models.Order, error)
}

type UserWithdrawalsFetcher interface {
	Call(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

type userOrdersFetcher struct {
	store storage.Storage
}

type userWithdrawalsFetcher struct {
	store storage.Storage
}

func NewUserOrdersFetcher(store storage.Storage) UserOrdersFetcher {
	return userOrdersFetcher{
		store: store,
	}
}

func NewUserWithdrawalsFetcher(store storage.Storage) UserWithdrawalsFetcher {
	return userWithdrawalsFetcher{
		store: store,
	}
}

func (f userOrdersFetcher) Call(ctx context.Context, userID int) ([]models.Order, error) {
	return f.store.UserOrders(ctx, userID)
}

func (f userWithdrawalsFetcher) Call(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return f.store.UserWithdrawals(ctx, userID)
}
