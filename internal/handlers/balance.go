package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type BalanceHandlers struct {
	store storage.Storage
}

func NewBalanceHandlers(store storage.Storage) BalanceHandlers {
	return BalanceHandlers{store: store}
}

func (bh BalanceHandlers) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	userID, _ := middlewares.UserIDFromContext(r.Context())
	balance, err := bh.store.FindBalanceByUserID(r.Context(), userID)

	var notFoundErr storage.ErrBalanceNotFound
	if err != nil && !errors.As(err, &notFoundErr) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(balance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
