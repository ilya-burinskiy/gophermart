package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ilya-burinskiy/gophermart/internal/auth"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/stretchr/testify/require"
)

type want struct {
	code        int
	response    string
	contentType string
}

func marshalJSON(val interface{}, t *testing.T) string {
	result, err := json.Marshal(val)
	require.NoError(t, err)

	return string(result)
}

func hashPassword(password string, t *testing.T) string {
	result, err := auth.HashPassword(password)
	require.NoError(t, err)

	return result
}

func generateAuthCookie(user models.User, t *testing.T) *http.Cookie {
	jwtStr, err := auth.BuildJWTString(user)
	require.NoError(t, err)

	return &http.Cookie{
		Name:     "jwt",
		Value:    jwtStr,
		MaxAge:   int(auth.TokenExp / time.Second),
		HttpOnly: true,
	}
}
