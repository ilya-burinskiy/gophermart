package models

type Balance struct {
	ID              int `json:"-"`
	UserID          int `json:"-"`
	CurrentAmount   int `json:"current"`
	WithdrawnAmount int `json:"withdrawn"`
}
