package models

import "time"

type Withdrawal struct {
	ID          int       `json:"-"`
	OrderNumber string    `json:"number"`
	UserID      int       `json:"-"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
