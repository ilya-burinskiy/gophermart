package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type UserHandlers struct {
	store storage.Storage
}

func NewUserHandlers(store storage.Storage) UserHandlers {
	return UserHandlers{store: store}
}

func UserRouter(store storage.Storage) chi.Router {
	handlers := UserHandlers{store: store}
	registerSrv := services.NewRegisterUserService(store)
	router := chi.NewRouter()
	router.Use(middleware.AllowContentType("application/json"))
	router.Post("/register", handlers.Register(registerSrv))

	return router
}

func (h UserHandlers) Register(registerSrv services.UserRegistrator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type payload struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		w.Header().Set("Content-Type", "application/json")
		var requestBody payload
		decoder := json.NewDecoder(r.Body)
		encoder := json.NewEncoder(w)

		err := decoder.Decode(&requestBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode("invalid request body")
			return
		}

		jwtStr, err := registerSrv.Call(r.Context(), requestBody.Login, requestBody.Password)
		if err != nil {
			var notUniqErr storage.ErrUserNotUniq
			if errors.As(err, &notUniqErr) {
				w.WriteHeader((http.StatusConflict))
				encoder.Encode(err.Error())
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(err.Error())
			return
		}

		setJWTCookie(w, jwtStr)
		w.WriteHeader(http.StatusOK)
	}
}
