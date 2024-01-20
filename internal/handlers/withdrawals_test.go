package handlers_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

type withdrawalCreatorMock struct{ mock.Mock }

func (m *withdrawalCreatorMock) Call(ctx context.Context, userID int, orderNumber string, sum int) (models.Withdrawal, error) {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Get(0).(models.Withdrawal), args.Error(1)
}

type withdrawalCreatorCallResult struct {
	returnValue models.Withdrawal
	err         error
}

func TestCreateWithdrawalHandler(t *testing.T) {
	type requestBody struct {
		Order string `json:"order"`
		Sum   int    `json:"sum"`
	}
	ctrl := gomock.NewController(t)
	storageMock := mocks.NewMockStorage(ctrl)

	router := chi.NewRouter()
	handlers := handlers.NewWithdrawalHanlers(storageMock)
	withdrawalCreatorMock := new(withdrawalCreatorMock)
	router.Use(middlewares.Authenticate)
	router.Post("/api/user/balance/withdraw", handlers.Create(withdrawalCreatorMock))
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	currentUser := models.User{ID: 1, Login: "login", EncryptedPassword: hashPassword("password", t)}
	currentUserAuthCookie := generateAuthCookie(currentUser, t)
	testCases := []struct {
		name                        string
		httpMethod                  string
		path                        string
		reqBody                     string
		authCookie                  *http.Cookie
		contentType                 string
		withdrawalCreatorCallResult withdrawalCreatorCallResult
		want                        want
	}{
		{
			name:        "responses with ok status",
			httpMethod:  http.MethodPost,
			path:        "/api/user/balance/withdraw",
			reqBody:     marshalJSON(requestBody{Order: "12345", Sum: 100}, t),
			authCookie:  currentUserAuthCookie,
			contentType: "application/json",
			withdrawalCreatorCallResult: withdrawalCreatorCallResult{
				returnValue: models.Withdrawal{
					ID:          1,
					OrderNumber: "12345",
					UserID:      currentUser.ID,
					Sum:         100,
					ProcessedAt: time.Now(),
				},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:        "responses with unauthorized status if user is not authenticated",
			httpMethod:  http.MethodPost,
			path:        "/api/user/balance/withdraw",
			authCookie:  &http.Cookie{},
			contentType: "application/json",
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{
			name:        "responses with payment required status if there is not enough amount",
			httpMethod:  http.MethodPost,
			path:        "/api/user/balance/withdraw",
			reqBody:     marshalJSON(requestBody{Order: "12345", Sum: 100}, t),
			authCookie:  currentUserAuthCookie,
			contentType: "application/json",
			withdrawalCreatorCallResult: withdrawalCreatorCallResult{
				err: services.ErrNotEnoughAmount,
			},
			want: want{
				code:        http.StatusPaymentRequired,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:        "responses with internal server error",
			httpMethod:  http.MethodPost,
			path:        "/api/user/balance/withdraw",
			authCookie:  currentUserAuthCookie,
			contentType: "application/json",
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withdrawalCreatorMockCall := withdrawalCreatorMock.
				On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(
					tc.withdrawalCreatorCallResult.returnValue,
					tc.withdrawalCreatorCallResult.err,
				)
			defer withdrawalCreatorMockCall.Unset()

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

type fetchWithdrawalsMock struct{ mock.Mock }

func (m *fetchWithdrawalsMock) Call(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Withdrawal), args.Error(1)
}

type fetchWithdrawalsCallResult struct {
	returnValue []models.Withdrawal
	err         error
}

func TestGetWithdrawalsHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	storageMock := mocks.NewMockStorage(ctrl)

	router := chi.NewRouter()
	handlers := handlers.NewWithdrawalHanlers(storageMock)
	fetchSrv := new(fetchWithdrawalsMock)
	router.Use(middlewares.Authenticate)
	router.Get("/api/user/withdrawals", handlers.Get(fetchSrv))
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	currentUser := models.User{ID: 1, Login: "login", EncryptedPassword: hashPassword("password", t)}
	currentUserAuthCookie := generateAuthCookie(currentUser, t)
	createdAt := time.Now()
	testCases := []struct {
		name                       string
		httpMethod                 string
		path                       string
		authCookie                 *http.Cookie
		contentType                string
		fetchWithdrawalsCallResult fetchWithdrawalsCallResult
		want                       want
	}{
		{
			name:        "responses with accepted status",
			httpMethod:  http.MethodGet,
			path:        "/api/user/withdrawals",
			authCookie:  currentUserAuthCookie,
			contentType: "application/json",
			fetchWithdrawalsCallResult: fetchWithdrawalsCallResult{
				returnValue: []models.Withdrawal{
					{ID: 1, OrderNumber: "123", UserID: currentUser.ID, Sum: 10, ProcessedAt: createdAt},
					{ID: 2, OrderNumber: "456", UserID: currentUser.ID, Sum: 20, ProcessedAt: createdAt},
				},
			},
			want: want{
				code: http.StatusOK,
				response: marshalJSON(
					[]models.Withdrawal{
						{ID: 1, OrderNumber: "123", UserID: currentUser.ID, Sum: 10, ProcessedAt: createdAt},
						{ID: 2, OrderNumber: "456", UserID: currentUser.ID, Sum: 20, ProcessedAt: createdAt},
					},
					t,
				),
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:        "responses with not authorized status if user is not authenticated",
			httpMethod:  http.MethodGet,
			path:        "/api/user/withdrawals",
			authCookie:  &http.Cookie{},
			contentType: "application/json",
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{
			name:        "responses with no content status if no withdrawals were made",
			httpMethod:  http.MethodGet,
			path:        "/api/user/withdrawals",
			authCookie:  currentUserAuthCookie,
			contentType: "application/json",
			fetchWithdrawalsCallResult: fetchWithdrawalsCallResult{
				returnValue: []models.Withdrawal{},
			},
			want: want{
				code:        http.StatusNoContent,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:        "responses with internal server error if error occured",
			httpMethod:  http.MethodGet,
			path:        "/api/user/withdrawals",
			authCookie:  currentUserAuthCookie,
			contentType: "application/json",
			fetchWithdrawalsCallResult: fetchWithdrawalsCallResult{
				err: errors.New("error"),
			},
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fetchSrvMockCall := fetchSrv.
				On("Call", mock.Anything, mock.Anything).
				Return(
					tc.fetchWithdrawalsCallResult.returnValue,
					tc.fetchWithdrawalsCallResult.err,
				)
			defer fetchSrvMockCall.Unset()

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
