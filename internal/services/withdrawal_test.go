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
)

func TestCall(t *testing.T) {
	ctx := context.TODO()
	user := models.User{ID: 1}
	currentAmount := 1000
	balance := models.Balance{ID: 1, UserID: user.ID, CurrentAmount: currentAmount}
	orderNumber := "1234"
	sum := 10

	t.Run("it creates withdrawal and updates balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		txMock := mocks.NewMockTx(ctrl)
		txMock.EXPECT().Commit(gomock.Any()).AnyTimes().Return(nil)
		txMock.EXPECT().Rollback(gomock.Any()).AnyTimes().Return(nil)

		storeMock := mocks.NewMockStorage(ctrl)
		storeMock.EXPECT().BeginTransaction(gomock.Any()).Return(txMock, nil)
		storeMock.EXPECT().CreateWithdrawal(ctx, user.ID, orderNumber, sum)
		storeMock.EXPECT().FindBalanceByUserID(ctx, user.ID).Return(balance, nil)
		storeMock.EXPECT().UpdateBalanceWithdrawnAmount(ctx, balance.ID, balance.WithdrawnAmount+sum).Return(nil)
		storeMock.EXPECT().UpdateBalanceCurrentAmount(ctx, balance.ID, currentAmount-sum).Return(nil)

		withdrawalCreator := services.NewWithdrawalCreator(storeMock)
		_, err := withdrawalCreator.Call(ctx, user.ID, orderNumber, sum)
		assert.NoError(t, err)
	})

	t.Run("it creates withdrawal and creates balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		txMock := mocks.NewMockTx(ctrl)
		txMock.EXPECT().Commit(gomock.Any()).AnyTimes().Return(nil)
		txMock.EXPECT().Rollback(gomock.Any()).AnyTimes().Return(nil)

		emptyBalance := models.Balance{ID: 1, UserID: user.ID}
		storeMock := mocks.NewMockStorage(ctrl)
		storeMock.EXPECT().BeginTransaction(gomock.Any()).Return(txMock, nil)
		storeMock.EXPECT().CreateWithdrawal(ctx, user.ID, orderNumber, sum)
		storeMock.EXPECT().FindBalanceByUserID(ctx, user.ID).Return(models.Balance{}, storage.ErrBalanceNotFound{})
		storeMock.EXPECT().CreateBalance(ctx, user.ID, 0).Return(emptyBalance, nil)

		withdrawalCreator := services.NewWithdrawalCreator(storeMock)
		_, err := withdrawalCreator.Call(ctx, user.ID, orderNumber, sum)
		assert.Error(t, err)
		assert.Equal(t, err, services.ErrNotEnoughAmount)
	})
}
