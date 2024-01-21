package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya-burinskiy/gophermart/internal/accrual"
	"github.com/ilya-burinskiy/gophermart/internal/configs"
	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"go.uber.org/zap"
)

func main() {
	config := configs.Parse()
	db, err := storage.NewDBStorage(config.DSN)
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

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM)

	configureUserRouter(db, router)
	configureOrderRouter(db, logger, config, exitCh, router)
	configureBalanceRouter(db, router)
	configureWithdrawalsRouter(db, router)

	server := http.Server{
		Handler: router,
		Addr:    config.RunAddr,
	}
	go onExit(db, &server, exitCh)
	server.ListenAndServe()
}

func onExit(store storage.Storage, srv *http.Server, exitCh <-chan os.Signal) {
	<-exitCh
	store.Close()
	srv.Shutdown(context.TODO())
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

func configureOrderRouter(
	store storage.Storage,
	logger *zap.Logger,
	config configs.Config,
	exitCh <-chan os.Signal,
	mainRouter chi.Router) {

	handlers := handlers.NewOrderHandlers(store)
	accrualApiClient := accrual.NewClient(config.AccrualBaseURL)
	accrualSrv := services.NewAccrualWorker(accrualApiClient, store, logger, exitCh)
	go accrualSrv.Run()
	createSrv := services.NewOrderCreateService(store, accrualSrv)
	fetchSrv := services.NewUserOrdersFetcher(store)
	mainRouter.Group(func(router chi.Router) {
		router.Use(
			middlewares.Authenticate,
			middleware.AllowContentType("text/plain"),
		)
		router.Post("/api/user/orders", handlers.Create(createSrv))
	})
	mainRouter.Group(func(router chi.Router) {
		router.Use(
			middlewares.Authenticate,
			middleware.AllowContentType("application/json"),
		)
		router.Get("/api/user/orders", handlers.Get(fetchSrv))
	})
}

func configureBalanceRouter(store storage.Storage, mainRouter chi.Router) {
	handlers := handlers.NewBalanceHandlers(store)
	mainRouter.Group(func(router chi.Router) {
		router.Use(
			middlewares.Authenticate,
			middleware.AllowContentType("application/json"),
		)
		router.Get("/api/user/balance", handlers.Get)
	})
}

func configureWithdrawalsRouter(store storage.Storage, mainRouter chi.Router) {
	handlers := handlers.NewWithdrawalHanlers(store)
	fetchSrv := services.NewUserWithdrawalsFetcher(store)
	createSrv := services.NewWithdrawalCreator(store)
	mainRouter.Group(func(router chi.Router) {
		router.Use(
			middlewares.Authenticate,
			middleware.AllowContentType("application/json"),
		)
		router.Get("/api/user/withdrawals", handlers.Get(fetchSrv))
		router.Post("/api/user/balance/withdraw", handlers.Create(createSrv))
	})
}
