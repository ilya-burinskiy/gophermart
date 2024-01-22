package services

import (
	"context"
	"errors"
	"os"

	"github.com/ilya-burinskiy/gophermart/internal/accrual"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AccrualWorker interface {
	Run()
	Register(order models.Order)
}

type accrualWorker struct {
	client  accrual.ApiClient
	store   storage.Storage
	logger  *zap.Logger
	channel chan models.Order
	exitCh  <-chan os.Signal
}

func NewAccrualWorker(
	accrualApiClient accrual.ApiClient,
	store storage.Storage,
	logger *zap.Logger,
	exitCh <-chan os.Signal) AccrualWorker {

	return accrualWorker{
		client:  accrualApiClient,
		store:   store,
		logger:  logger,
		exitCh:  exitCh,
		channel: make(chan models.Order),
	}
}

func (wrk accrualWorker) Register(order models.Order) {
	wrk.channel <- order
}

func (wrk accrualWorker) Run() {
Loop:
	for {
		select {
		case order := <-wrk.channel:
			orderInfo, err := wrk.client.GetOrderInfo(context.TODO(), order.Number)
			if err != nil {
				wrk.logger.Info(
					"failed to get order info",
					zap.String("order_number", order.Number),
					zap.Error(err),
				)
				continue
			}

			ctx := context.TODO()
			wrk.updateOrderWithBalance(ctx, order, orderInfo)
		case <-wrk.exitCh:
			break Loop
		}
	}
}

func (wrk accrualWorker) updateOrderWithBalance(ctx context.Context, order models.Order, orderInfo accrual.OrderInfo) {
	wrk.store.WithinTranscaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		err := wrk.store.UpdateOrderTx(ctx, tx, order.ID, orderInfo.Status, orderInfo.Accrual)
		if err != nil {
			wrk.logger.Info("failed to update order", zap.Error(err))
			return err
		}

		balance, err := wrk.store.FindBalanceByUserIDTx(ctx, tx, order.UserID)
		if err != nil {
			var notFoundErr storage.ErrBalanceNotFound
			if errors.As(err, &notFoundErr) {
				balance, err = wrk.store.CreateBalanceTx(ctx, tx, order.UserID, orderInfo.Accrual)
				if err != nil {
					wrk.logger.Info("failed to create balance", zap.Error(err))
					return err
				}
			}

			wrk.logger.Info("an unexpted error occured while trying to find balance", zap.Error(err))
			return err
		}

		err = wrk.store.UpdateBalanceCurrentAmountTx(ctx, tx, balance.ID, balance.CurrentAmount+orderInfo.Accrual)
		if err != nil {
			wrk.logger.Info("failed to updage balance current amount", zap.Error(err))
			return err
		}

		return nil
	})
}
