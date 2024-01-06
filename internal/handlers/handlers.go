package handlers

import (
	"net/http"
	"time"

	"github.com/ilya-burinskiy/gophermart/internal/auth"
)

func setJWTCookie(w http.ResponseWriter, token string) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "jwt",
			Value:    token,
			MaxAge:   int(auth.TokenExp / time.Second),
			HttpOnly: true,
		},
	)
}
