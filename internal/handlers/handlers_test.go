package handlers_test

import (
	"encoding/json"
	"testing"

	"github.com/ilya-burinskiy/gophermart/internal/auth"
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
