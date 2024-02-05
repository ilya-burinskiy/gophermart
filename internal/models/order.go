package models

import (
	"encoding/json"
	"time"
)

type OrderStatus int

const (
	NewOrder OrderStatus = iota
	RegisteredOrder
	ProcessingOrder
	InvalidOrder
	ProcessedOrder
)

type Order struct {
	ID        int         `json:"-"`
	UserID    int         `json:"-"`
	Number    string      `json:"number"`
	Status    OrderStatus `json:"status"`
	Accrual   int         `json:"accrual"`
	CreatedAt time.Time   `json:"uploaded_at"`
}

func (order Order) MarshalJSON() ([]byte, error) {
	type OrderAlias Order

	orderStatus2String := map[OrderStatus]string{
		NewOrder:        "NEW",
		RegisteredOrder: "REGISTERED",
		ProcessingOrder: "PROCESSING",
		ProcessedOrder:  "PROCESSED",
		InvalidOrder:    "INVALID",
	}
	aliasValue := struct {
		OrderAlias
		Status string `json:"status"`
	}{
		OrderAlias: OrderAlias(order),
		Status:     orderStatus2String[order.Status],
	}

	return json.Marshal(aliasValue)
}
