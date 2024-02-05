package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/ilya-burinskiy/gophermart/internal/auth"
	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"github.com/ilya-burinskiy/gophermart/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type userRegistratorMock struct{ mock.Mock }

func (m *userRegistratorMock) Call(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

type userRegistratorCallResult struct {
	returnValue string
	err         error
}

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	// TODO: maybe use stub instead of mock
	storageMock := mocks.NewMockStorage(ctrl)
	userRegistratorMock := new(userRegistratorMock)

	router := chi.NewRouter()
	handlers := handlers.NewUserHandlers(storageMock)
	router.Post("/api/user/register", handlers.Register(userRegistratorMock))

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	testCases := []struct {
		name                      string
		httpMethod                string
		path                      string
		reqBody                   string
		contentType               string
		userRegistratorCallResult userRegistratorCallResult
		want                      want
	}{
		{
			name:        "responses with ok status",
			httpMethod:  http.MethodPost,
			path:        "/api/user/register",
			reqBody:     marshalJSON(map[string]string{"login": "login", "password": "password"}, t),
			contentType: "application/json",
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name:        "responses with bad request status if could not parse json body",
			httpMethod:  http.MethodPost,
			path:        "/api/user/register",
			reqBody:     marshalJSON("body", t),
			contentType: "application/json",
			want: want{
				code:        http.StatusBadRequest,
				response:    "\"invalid request body\"\n",
				contentType: "application/json",
			},
		},
		{
			name:        "responses with conflict status if user already registered",
			httpMethod:  http.MethodPost,
			path:        "/api/user/register",
			reqBody:     marshalJSON(map[string]string{"login": "login", "password": "password"}, t),
			contentType: "application/json",
			userRegistratorCallResult: userRegistratorCallResult{
				err:         storage.ErrUserNotUniq{User: models.User{ID: 1, Login: "login"}},
			},
			want: want{
				code:        http.StatusConflict,
				response:    "\"user with login \\\"login\\\" already exists\"\n",
				contentType: "application/json",
			},
		},
		{
			name:        "responses with internal server error status if could not register user",
			httpMethod:  http.MethodPost,
			path:        "/api/user/register",
			reqBody:     marshalJSON(map[string]string{"login": "login", "password": "password"}, t),
			contentType: "application/json",
			userRegistratorCallResult: userRegistratorCallResult{
				err:         fmt.Errorf("failed to generate JWT"),
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "\"failed to generate JWT\"\n",
				contentType: "application/json",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registerUserMockCall := userRegistratorMock.
				On("Call", mock.Anything, mock.Anything, mock.Anything).
				Return(
					tc.userRegistratorCallResult.returnValue,
					tc.userRegistratorCallResult.err,
				)
			defer registerUserMockCall.Unset()

			request, err := http.NewRequest(
				tc.httpMethod,
				testServer.URL+tc.path,
				strings.NewReader(tc.reqBody),
			)
			require.NoError(t, err)
			request.Header.Set("Content-Type", tc.contentType)
			request.Header.Set("Accept-Encoding", "identity")

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

type userAuthenticatorMock struct{ mock.Mock }

func (m *userAuthenticatorMock) Call(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

type userAuthenticatorCallResult struct {
	returnValue string
	err         error
}

func TestAuthenticateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	// TODO: maybe use stub instead of mock
	storageMock := mocks.NewMockStorage(ctrl)
	userAuthenticatorMock := new(userAuthenticatorMock)

	router := chi.NewRouter()
	handlers := handlers.NewUserHandlers(storageMock)
	router.Post("/api/user/login", handlers.Authenticate(userAuthenticatorMock))

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	testCases := []struct {
		name                        string
		httpMethod                  string
		path                        string
		reqBody                     string
		contentType                 string
		userAuthenticatorCallResult userAuthenticatorCallResult
		want                        want
	}{
		{
			name:       "responses with ok status",
			httpMethod: http.MethodPost,
			path:       "/api/user/login",
			reqBody: marshalJSON(
				map[string]string{
					"login":    "login",
					"password": "password",
				},
				t,
			),
			contentType: "application/json",
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name:        "responses with bad request status if could not parse json body",
			httpMethod:  http.MethodPost,
			path:        "/api/user/login",
			reqBody:     marshalJSON("body", t),
			contentType: "application/json",
			want: want{
				code:        http.StatusBadRequest,
				response:    "\"invalid request body\"\n",
				contentType: "application/json",
			},
		},
		{
			name:       "responses with not authorized status if could not find user by login",
			httpMethod: http.MethodPost,
			path:       "/api/user/login",
			reqBody: marshalJSON(
				map[string]string{
					"login":    "login",
					"password": "password",
				},
				t,
			),
			contentType: "application/json",
			userAuthenticatorCallResult: userAuthenticatorCallResult{
				err: fmt.Errorf(
					"failed to authenticate user: %w",
					storage.ErrUserNotFound{User: models.User{ID: 1, Login: "login"}},
				),
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    "\"failed to authenticate user: user with login \\\"login\\\" not found\"\n",
				contentType: "application/json",
			},
		},
		{
			name:       "responses with not authorized status if login or password are invalid",
			httpMethod: http.MethodPost,
			path:       "/api/user/login",
			reqBody: marshalJSON(
				map[string]string{
					"login":    "login",
					"password": "password",
				},
				t,
			),
			contentType: "application/json",
			userAuthenticatorCallResult: userAuthenticatorCallResult{
				err: fmt.Errorf(
					"failed to authenticate user: %w",
					auth.ErrInvalidCreds,
				),
			},
			want: want{
				code:        http.StatusUnauthorized,
				response:    "\"failed to authenticate user: invalid login or password\"\n",
				contentType: "application/json",
			},
		},
		{
			name:       "responses with internal server error status if an error occured",
			httpMethod: http.MethodPost,
			path:       "/api/user/login",
			reqBody: marshalJSON(
				map[string]string{
					"login":    "login",
					"password": "password",
				},
				t,
			),
			contentType: "application/json",
			userAuthenticatorCallResult: userAuthenticatorCallResult{
				err: fmt.Errorf("failed to authenticate user: jwt error"),
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "\"failed to authenticate user: jwt error\"\n",
				contentType: "application/json",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authenticateMockCall := userAuthenticatorMock.
				On("Call", mock.Anything, mock.Anything, mock.Anything).
				Return(
					tc.userAuthenticatorCallResult.returnValue,
					tc.userAuthenticatorCallResult.err,
				)
			defer authenticateMockCall.Unset()

			request, err := http.NewRequest(
				tc.httpMethod,
				testServer.URL+tc.path,
				strings.NewReader(tc.reqBody),
			)
			require.NoError(t, err)
			request.Header.Set("Content-Type", tc.contentType)
			request.Header.Set("Accept-Encoding", "identity")

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
