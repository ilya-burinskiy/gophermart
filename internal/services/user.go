package services

import (
	"context"

	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type UserOrdersFetcher interface {
	Call(ctx context.Context, userID int) ([]models.Order, error)
}

type userOrdersFetcher struct {
	store storage.Storage
}

func NewUserOrdersFetcher(store storage.Storage) UserOrdersFetcher {
	return userOrdersFetcher{
		store: store,
	}
}

func (f userOrdersFetcher) Call(ctx context.Context, userID int) ([]models.Order, error) {
	return f.store.UserOrders(ctx, userID)
}
