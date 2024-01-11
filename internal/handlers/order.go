package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"go.uber.org/zap"
)

type orderHandlers struct {
	store  storage.Storage
	logger *zap.Logger
}

func OrderRouter(store storage.Storage, logger *zap.Logger) chi.Router {
	router := chi.NewRouter()
	handlers := orderHandlers{store: store, logger: logger}
	router.Use(
		middlewares.Authenticate,
		middleware.AllowContentType("text/plain"),
	)
	router.Post("/", handlers.Create)

	return router
}

func (oh orderHandlers) Create(w http.ResponseWriter, r *http.Request) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		oh.logger.Info(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	orderNumber := string(rawBody)
	userID, _ := middlewares.UserIDFromContext(r.Context())
	order, err := oh.store.CreateOrder(
		r.Context(),
		userID,
		orderNumber,
		models.NewOrder,
	)

	if err != nil {
		oh.logger.Info(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	oh.logger.Info(fmt.Sprintf("%v", order))
	w.WriteHeader(http.StatusAccepted)
}
