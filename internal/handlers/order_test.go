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
	"github.com/ilya-burinskiy/gophermart/internal/handlers"
	"github.com/ilya-burinskiy/gophermart/internal/middlewares"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type orderCreaterMock struct{ mock.Mock }

func (m *orderCreaterMock) Call(ctx context.Context, number string, userID int) (models.Order, error) {
	args := m.Called(ctx, number, userID)
	return args.Get(0).(models.Order), args.Error(1)
}

type orderCreaterCallResult struct {
	returnValue models.Order
	err         error
}

func TestCreateOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	// TODO: maybe use stub instead of mock
	storageMock := mocks.NewMockStorage(ctrl)

	router := chi.NewRouter()
	handlers := handlers.NewOrderHandlers(storageMock)
	createSrvMock := new(orderCreaterMock)
	router.Use(middlewares.Authenticate)
	router.Post("/api/user/orders", handlers.Create(createSrvMock))
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	currentUser := models.User{ID: 1, Login: "login", EncryptedPassword: hashPassword("password", t)}
	testCases := []struct {
		name                   string
		httpMethod             string
		path                   string
		reqBody                string
		authCookie             *http.Cookie
		contentType            string
		orderCreaterCallResult orderCreaterCallResult
		want                   want
	}{
		{
			name:        "responses with accepted status",
			httpMethod:  http.MethodPost,
			path:        "/api/user/orders",
			reqBody:     "12345",
			authCookie:  generateAuthCookie(currentUser, t),
			contentType: "text/plain",
			want: want{
				code:        http.StatusAccepted,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "responses with bad request status if request is invalid",
			httpMethod:  http.MethodPost,
			path:        "/api/user/orders",
			reqBody:     "",
			authCookie:  generateAuthCookie(currentUser, t),
			contentType: "text/plain",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "responses with unauthorized status if user is not authenticated",
			httpMethod:  http.MethodPost,
			path:        "/api/user/orders",
			reqBody:     "12345",
			authCookie:  &http.Cookie{},
			contentType: "text/plain",
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{
			name:        "responses with ok status if order was already uploaded by current user",
			httpMethod:  http.MethodPost,
			path:        "/api/user/orders",
			reqBody:     "12345",
			authCookie:  generateAuthCookie(currentUser, t),
			contentType: "text/plain",
			orderCreaterCallResult: orderCreaterCallResult{
				err: services.ErrDuplicatedOrder{
					Order: models.Order{ID: 1, UserID: 1, Number: "12345"},
				},
			},
			want: want{
				code:        http.StatusOK,
				response:    "order with number \"12345\" already exists",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "responses with conflict status if order was already uploaded by another user",
			httpMethod:  http.MethodPost,
			path:        "/api/user/orders",
			reqBody:     "12345",
			authCookie:  generateAuthCookie(currentUser, t),
			contentType: "text/plain",
			orderCreaterCallResult: orderCreaterCallResult{
				err: services.ErrConflicOrder{
					Order: models.Order{ID: 1, UserID: 2, Number: "12345"},
				},
			},
			want: want{
				code:        http.StatusConflict,
				response:    "order with number \"12345\" was created by another user",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "responses with internal server status if error occured",
			httpMethod:  http.MethodPost,
			path:        "/api/user/orders",
			reqBody:     "12345",
			authCookie:  generateAuthCookie(currentUser, t),
			contentType: "text/plain",
			orderCreaterCallResult: orderCreaterCallResult{
				err: fmt.Errorf(
					"failed to create order: error",
				),
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "failed to create order: error",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createSrvMockCall := createSrvMock.
				On("Call", mock.Anything, mock.Anything, mock.Anything).
				Return(
					tc.orderCreaterCallResult.returnValue,
					tc.orderCreaterCallResult.err,
				)
			defer createSrvMockCall.Unset()

			request, err := http.NewRequest(
				tc.httpMethod,
				testServer.URL+tc.path,
				strings.NewReader(tc.reqBody),
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
