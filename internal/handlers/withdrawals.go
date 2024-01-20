package handlers

import (
	"encoding/json"
	"errors"
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

func (wh WithdrawalHandlers) Create(createSrv services.WithdrawalCreator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type payload struct {
			Order string `json:"order"`
			Sum   int    `json:"sum"`
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		decoder := json.NewDecoder(r.Body)
		var requestBody payload
		err := decoder.Decode(&requestBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userID, _ := middlewares.UserIDFromContext(r.Context())
		_, err = createSrv.Call(r.Context(), userID, requestBody.Order, requestBody.Sum)
		if err != nil {
			if errors.Is(err, services.ErrNotEnoughAmount) {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
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
