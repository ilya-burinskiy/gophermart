package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type OrderCreater interface {
	Call(ctx context.Context, number string, userID int) (models.Order, error)
}

type OrderCreateService struct {
	store storage.Storage
}

func NewOrderCreateService(store storage.Storage) OrderCreateService {
	return OrderCreateService{
		store: store,
	}
}

type ErrDuplicatedOrder struct {
	Order models.Order
}

func (err ErrDuplicatedOrder) Error() string {
	return fmt.Sprintf("order with number \"%s\" already exists", err.Order.Number)
}

type ErrConflicOrder struct {
	Order models.Order
}

func (err ErrConflicOrder) Error() string {
	return fmt.Sprintf("order with number \"%s\" was created by another user", err.Order.Number)
}

func (srv OrderCreateService) Call(ctx context.Context, number string, userID int) (models.Order, error) {
	// TODO: add number validation
	order, err := srv.store.CreateOrder(
		ctx,
		userID,
		number,
		models.NewOrder,
	)

	if err != nil {
		var notUniqErr storage.ErrOrderNotUnique
		if errors.As(err, &notUniqErr) {
			existingOrder, err := srv.store.FindOrderByNumber(ctx, number)
			if err != nil {
				// TODO: better error msg
				return models.Order{}, fmt.Errorf("failed to create order: %w", err)
			}

			if existingOrder.UserID == userID {
				return models.Order{}, ErrDuplicatedOrder{Order: order}
			}

			return models.Order{}, ErrConflicOrder{Order: existingOrder}
		}

		return models.Order{}, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}
