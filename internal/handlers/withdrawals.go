package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type WithdrawalHandlers struct {
	store storage.Storage
}

func NewWithdrawalHanlers(store storage.Storage) WithdrawalHandlers {
	return WithdrawalHandlers{store: store}
}

func (wh WithdrawalHandlers) Get(fetchSrv services.UserWithdrawalsFetcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		userID, _ := middlewares.UserIDFromContext(r.Context())
		withdrawals, err := fetchSrv.Call(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		responseBody, err := json.Marshal(withdrawals)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	}
}
