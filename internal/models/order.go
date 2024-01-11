package models

import "time"

type OrderStatus int

const (
	NewOrder OrderStatus = iota
	ProcessingOrder
	InvalidOrder
	ProcessedOrder
)

type Order struct {
	ID        int
	UserID    int
	Number    string
	Status    OrderStatus
	Accrual   int
	CreatedAt time.Time
}
