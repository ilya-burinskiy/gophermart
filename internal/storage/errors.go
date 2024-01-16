package storage

import (
	"fmt"

	"github.com/ilya-burinskiy/gophermart/internal/models"
)

type ErrUserNotUniq struct {
	User models.User
}

func (err ErrUserNotUniq) Error() string {
	return fmt.Sprintf("user with login \"%s\" already exists", err.User.Login)
}

type ErrOrderNotUnique struct {
	order models.Order
}

func (err ErrOrderNotUnique) Error() string {
	return fmt.Sprintf("order with number \"%s\" already exists", err.order.Number)
}

type ErrUserNotFound struct {
	User models.User
}

func (err ErrUserNotFound) Error() string {
	return fmt.Sprintf("user with login \"%s\" not found", err.User.Login)
}

type ErrOrderNotFound struct {
	Order models.Order
}

func (err ErrOrderNotFound) Error() string {
	return fmt.Sprintf("order with number \"%s\" not found", err.Order.Number)
}

type ErrBalanceNotUnique struct {
	Balance models.Balance
}

func (err ErrBalanceNotUnique) Error() string {
	return fmt.Sprintf("balance with \"user_id\"=%d already exists", err.Balance.UserID)
}

type ErrBalanceNotFound struct {
	Balance models.Balance
}

func (err ErrBalanceNotFound) Error() string {
	return fmt.Sprintf("balance with \"user_id\"=%d not found", err.Balance.UserID)
}
