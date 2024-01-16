package models

type Balance struct {
	ID              int
	UserID          int
	CurrentAmount   int
	WithdrawnAmount int
}
