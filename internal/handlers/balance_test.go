package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBalanceHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	storageMock := mocks.NewMockStorage(ctrl)

	currentUser := models.User{ID: 1, Login: "login", EncryptedPassword: hashPassword("password", t)}
	balance := models.Balance{ID: 1, UserID: currentUser.ID}
	storageMock.EXPECT().
		FindBalanceByUserID(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(balance, nil)

	router := chi.NewRouter()
	handlers := handlers.NewBalanceHandlers(storageMock)
	router.Use(middlewares.Authenticate)
	router.Get("/api/user/balance", handlers.Get)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	currentUserAuthCookie := generateAuthCookie(currentUser, t)
	testCases := []struct {
		name        string
		httpMethod  string
		path        string
		authCookie  *http.Cookie
		contentType string
		want        want
	}{
		{
			name:        "responses with ok status",
			httpMethod:  http.MethodGet,
			path:        "/api/user/balance",
			authCookie:  currentUserAuthCookie,
			contentType: "application/json",
			want: want{
				code:        http.StatusOK,
				response:    marshalJSON(balance, t),
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:        "responses with unauthorized status if user is not authenticated",
			httpMethod:  http.MethodGet,
			path:        "/api/user/balance",
			authCookie:  &http.Cookie{},
			contentType: "application/json",
			want: want{
				code: http.StatusUnauthorized,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request, err := http.NewRequest(
				tc.httpMethod,
				testServer.URL+tc.path,
				nil,
			)
			require.NoError(t, err)
			request.Header.Set("Content-Type", tc.contentType)
			request.Header.Set("Accept-Encoding", "identity")
			request.AddCookie(tc.authCookie)

			response, err := testServer.Client().Do(request)
			require.NoError(t, err)
			resBody, err := io.ReadAll(response.Body)
			defer response.Body.Close()

			assert.NoError(t, err)
			assert.Equal(t, tc.want.code, response.StatusCode)
			assert.Equal(t, tc.want.response, string(resBody))
			assert.Equal(t, tc.want.contentType, response.Header.Get("Content-Type"))
		})
	}
}
