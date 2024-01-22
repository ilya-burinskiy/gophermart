package services_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ilya-burinskiy/gophermart/internal/accrual"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage/mocks"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type accrualApiClientMock struct{ mock.Mock }

func (m *accrualApiClientMock) GetOrderInfo(ctx context.Context, orderNumber string) (accrual.OrderInfo, error) {
	call := m.Called(ctx, orderNumber)
	return call.Get(0).(accrual.OrderInfo), call.Error(1)
}

func TestAccrualRun(t *testing.T) {
	orderInfo := accrual.OrderInfo{Number: "123", Status: models.ProcessedOrder, Accrual: 10}
	order := models.Order{ID: 1, UserID: 1, Number: "123"}
	exitCh := make(chan os.Signal)

	logger := zaptest.NewLogger(t)
	accrualApiClientMock := new(accrualApiClientMock)
	accrualApiClientMock.
		On("GetOrderInfo", mock.Anything, "123").
		Return(orderInfo, nil)

	t.Run("it updates order and balance current amount", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		storageMock := mocks.NewMockStorage(ctrl)
		// TODO: write expectations
		accrualWrk := services.NewAccrualWorker(accrualApiClientMock, storageMock, logger, exitCh)
		go accrualWrk.Run()
		accrualWrk.Register(order)
	})

	t.Run("it creates balance and updates order", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		storageMock := mocks.NewMockStorage(ctrl)
		// TODO: write expectations
		accrualWrk := services.NewAccrualWorker(accrualApiClientMock, storageMock, logger, exitCh)
		go accrualWrk.Run()
		accrualWrk.Register(order)
	})
}
