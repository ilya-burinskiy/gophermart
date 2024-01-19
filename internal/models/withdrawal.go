package models

import "time"

type Withdrawal struct {
	ID          int
	OrderNumber string
	UserID      int
	Sum         int
	ProcessedAt time.Time
}
