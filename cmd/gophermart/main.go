package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"go.uber.org/zap"
)

func main() {
	// TODO: make server addr and DSN configurable
	db, err := storage.NewDBStorage("postgres://gophermart:password@localhost:5432/gophermart")
	if err != nil {
		panic(err)
	}
	logger := configureLogger("info")

	router := chi.NewRouter()
	router.Use(
		middlewares.LogResponse(logger),
		middlewares.LogRequest(logger),
		middlewares.GzipCompress,
		middleware.AllowContentEncoding("gzip"),
	)
	configureUserRouter(db, router)
	configureOrderRouter(db, logger, router)

	server := http.Server{
		Handler: router,
		Addr:    "localhost:8080",
	}
	server.ListenAndServe()
}

func configureLogger(level string) *zap.Logger {
	logLvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		panic(err)
	}
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = logLvl
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

func configureUserRouter(store storage.Storage, mainRouter chi.Router) {
	handlers := handlers.NewUserHandlers(store)
	registerSrv := services.NewRegisterUserService(store)
	authenticateSrv := services.NewAuthenticateUserService(store)

	mainRouter.Group(func(router chi.Router) {
		router.Use(middleware.AllowContentType("application/json"))
		router.Post("/api/user/register", handlers.Register(registerSrv))
		router.Post("/api/user/login", handlers.Authenticate(authenticateSrv))
	})
}

func configureOrderRouter(store storage.Storage, logger *zap.Logger, mainRouter chi.Router) {
	handlers := handlers.NewOrderHandlers(store, logger)
	mainRouter.Group(func(router chi.Router) {
		router.Use(
			middlewares.Authenticate,
			middleware.AllowContentType("text/plain"),
		)
		router.Post("/api/user/orders", handlers.Create)
	})
}
