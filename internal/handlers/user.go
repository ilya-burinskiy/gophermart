package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ilya-burinskiy/gophermart/internal/auth"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
)

type UserHandlers struct {
	store storage.Storage
}

func NewUserHandlers(store storage.Storage) UserHandlers {
	return UserHandlers{store: store}
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

		auth.SetJWTCookie(w, jwtStr)
		w.WriteHeader(http.StatusOK)
	}
}

func (h UserHandlers) Authenticate(authSrv services.UserAuthenticator) func(http.ResponseWriter, *http.Request) {
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

		jwtStr, err := authSrv.Call(r.Context(), requestBody.Login, requestBody.Password)
		if err != nil {
			var notFoundErr storage.ErrUserNotFound
			fmt.Println(err.Error())
			if errors.Is(err, auth.ErrInvalidCreds) || errors.As(err, &notFoundErr) {
				w.WriteHeader(http.StatusUnauthorized)
				encoder.Encode(err.Error())
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(err.Error())
			return
		}

		auth.SetJWTCookie(w, jwtStr)
		w.WriteHeader(http.StatusOK)
	}
}
