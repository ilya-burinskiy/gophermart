package services_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/services"
	"github.com/ilya-burinskiy/gophermart/internal/storage/mocks"
)

func TestCall(t *testing.T) {
	ctx := context.TODO()
	user := models.User{ID: 1}
	orderNumber := "1234"
	sum := 10

	t.Run("it creates withdrawal and updates balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		txMock := mocks.NewMockTx(ctrl)
		txMock.EXPECT().Commit(gomock.Any()).AnyTimes().Return(nil)
		txMock.EXPECT().Rollback(gomock.Any()).AnyTimes().Return(nil)

		storeMock := mocks.NewMockStorage(ctrl)
		// TODO: write expectations
		withdrawalCreator := services.NewWithdrawalCreator(storeMock)
		withdrawalCreator.Call(ctx, user.ID, orderNumber, sum)
	})

	t.Run("it creates withdrawal and creates balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		storeMock := mocks.NewMockStorage(ctrl)
		// TODO: write expectations
		withdrawalCreator := services.NewWithdrawalCreator(storeMock)
		withdrawalCreator.Call(ctx, user.ID, orderNumber, sum)
	})
}
