package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"github.com/ilya-burinskiy/gophermart/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type want struct {
	code        int
	response    string
	contentType string
}

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
	storageMock := mocks.NewMockStorage(ctrl)
	storageMock.EXPECT().
		CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(models.User{ID: 1, Login: "login", EncryptedPassword: "abcd"}, nil)
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
				response:    "",
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
				returnValue: "",
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
				returnValue: "",
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

func marshalJSON(val interface{}, t *testing.T) string {
	result, err := json.Marshal(val)
	require.NoError(t, err)

	return string(result)
}
