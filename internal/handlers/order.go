package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type OrderHandlers struct {
	store storage.Storage
}

func NewOrderHandlers(store storage.Storage) OrderHandlers {
	return OrderHandlers{store: store}
}

func (oh OrderHandlers) Create(createSrv services.OrderCreater) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rawBody, err := io.ReadAll(r.Body)
		if err != nil || len(rawBody) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		orderNumber := string(rawBody)
		userID, _ := middlewares.UserIDFromContext(r.Context())
		_, err = createSrv.Call(r.Context(), orderNumber, userID)

		var duplicateErr services.ErrDuplicatedOrder
		var conflictErr services.ErrConflicOrder
		if err != nil {
			switch {
			case errors.As(err, &duplicateErr):
				w.WriteHeader(http.StatusOK)
			case errors.As(err, &conflictErr):
				w.WriteHeader(http.StatusConflict)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (oh OrderHandlers) Get(fetchSrv services.UserOrdersFetcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		userID, _ := middlewares.UserIDFromContext(r.Context())
		orders, err := fetchSrv.Call(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		responseBody, err := json.Marshal(orders)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	}
}
