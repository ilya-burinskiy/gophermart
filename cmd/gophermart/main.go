package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
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
	router.Mount("/api/user", handlers.UserRouter(db))
	router.Mount("/aip/user/orders", handlers.OrderRouter(db, logger))

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
