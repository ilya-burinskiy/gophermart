package services_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"github.com/ilya-burinskiy/gophermart/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type accrualWorkerMock struct{ mock.Mock }

func (m *accrualWorkerMock) Run()                        {}
func (m *accrualWorkerMock) Register(order models.Order) {}

func TestCreateOrderCall(t *testing.T) {
	accrualWorker := new(accrualWorkerMock)
	owner := models.User{ID: 1}
	user := models.User{ID: 2}
	orderNumber := "1234"
	ctx := context.TODO()

	t.Run("it creates order", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		storageMock := mocks.NewMockStorage(ctrl)
		createSrv := services.NewOrderCreateService(storageMock, accrualWorker)

		storageMock.EXPECT().CreateOrder(ctx, owner.ID, orderNumber, models.RegisteredOrder).
			Return(models.Order{ID: 1, UserID: owner.ID, Number: orderNumber}, nil)
		_, err := createSrv.Call(ctx, orderNumber, owner.ID)
		assert.NoError(t, err)
	})

	t.Run("it returns duplicate error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		storageMock := mocks.NewMockStorage(ctrl)
		createSrv := services.NewOrderCreateService(storageMock, accrualWorker)
		existingOrder := models.Order{ID: 1, Number: orderNumber, UserID: owner.ID}

		storageMock.EXPECT().CreateOrder(ctx, owner.ID, orderNumber, models.RegisteredOrder).
			Return(existingOrder, storage.ErrOrderNotUnique{})
		storageMock.EXPECT().FindOrderByNumber(ctx, orderNumber).
			Return(existingOrder, nil)

		_, err := createSrv.Call(ctx, orderNumber, owner.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrDuplicatedOrder{Order: existingOrder}, err)
	})

	t.Run("it returns conflict error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		storageMock := mocks.NewMockStorage(ctrl)
		createSrv := services.NewOrderCreateService(storageMock, accrualWorker)
		existingOrder := models.Order{ID: 1, Number: orderNumber, UserID: user.ID}

		storageMock.EXPECT().CreateOrder(ctx, owner.ID, orderNumber, models.RegisteredOrder).
			Return(existingOrder, storage.ErrOrderNotUnique{})
		storageMock.EXPECT().FindOrderByNumber(ctx, orderNumber).
			Return(existingOrder, nil)

		_, err := createSrv.Call(ctx, orderNumber, owner.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrConflicOrder{Order: existingOrder}, err)
	})
}
