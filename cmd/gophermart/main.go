package main

import (
	"net/http"

	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

func main() {
	// TODO: make server addr and DSN configurable
	db, err := storage.NewDBStorage("postgres://gophermart:password@localhost:5432/gophermart")
	if err != nil {
		panic(err)
	}
	router := handlers.UserRouter(db)

	server := http.Server{
		Handler: router,
		Addr:    "localhost:8080",
	}
	server.ListenAndServe()
}
