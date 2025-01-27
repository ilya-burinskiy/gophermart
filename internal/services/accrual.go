package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ilya-burinskiy/gophermart/internal/accrual"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AccrualWorker interface {
	Run()
}

type accrualWorker struct {
	client     accrual.ApiClient
	store      storage.Storage
	logger     *zap.Logger
	workersNum int
	exitCh     <-chan struct{}
}

func NewAccrualWorker(
	accrualApiClient accrual.ApiClient,
	store storage.Storage,
	logger *zap.Logger,
	workersNum int,
	exitCh <-chan struct{}) AccrualWorker {

	return accrualWorker{
		client:     accrualApiClient,
		store:      store,
		logger:     logger,
		workersNum: workersNum,
		exitCh:     exitCh,
	}
}

func (wrk accrualWorker) Run() {
	jobsChannel := make(chan models.Order, wrk.workersNum)
	ticker := time.NewTicker(5 * time.Second)
	ctx := context.TODO()

	for w := 1; w <= wrk.workersNum; w++ {
		go wrk.processOrder(jobsChannel)
	}

	for {
		select {
		case <-ticker.C:
			orders, err := wrk.store.NewOrders(ctx)
			if err != nil {
				wrk.logger.Info("run accrual worker", zap.Error(err))
				continue
			}

			for _, order := range orders {
				jobsChannel <- order
			}
		case <-wrk.exitCh:
			wrk.logger.Info("finishing accrual worker")
			close(jobsChannel)
			return
		}
	}
}

func (wrk accrualWorker) processOrder(jobsChannel <-chan models.Order) {
	ctx := context.TODO()
	for order := range jobsChannel {
		orderInfo, err := wrk.client.GetOrderInfo(ctx, order.Number)
		if err != nil {
			wrk.logger.Info("accrual worker error", zap.Error(err))
			err = wrk.store.DeleteOrder(ctx, order.ID)
			if err != nil {
				wrk.logger.Info("accrual worker error", zap.Error(err))
			}
			continue
		}
		err = wrk.updateOrderWithBalance(ctx, order, orderInfo)
		if err != nil {
			wrk.logger.Info("accrual worker error", zap.Error(err))
		}
	}
}

func (wrk accrualWorker) updateOrderWithBalance(
	ctx context.Context,
	order models.Order,
	orderInfo accrual.OrderInfo) error {

	err := wrk.store.WithinTranscaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		err := wrk.store.UpdateOrderTx(ctx, tx, order.ID, orderInfo.Status, orderInfo.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		balance, err := wrk.store.FindBalanceByUserIDTx(ctx, tx, order.UserID)
		if err != nil {
			var notFoundErr storage.ErrBalanceNotFound
			if errors.As(err, &notFoundErr) {
				balance, err = wrk.store.CreateBalanceTx(ctx, tx, order.UserID, orderInfo.Accrual)
				if err != nil {
					return fmt.Errorf("failed to create balance: %w", err)
				}
			}

			return fmt.Errorf("an unexpted error occured while trying to find balance: %w", err)
		}

		err = wrk.store.UpdateBalanceCurrentAmountTx(ctx, tx, balance.ID, balance.CurrentAmount+orderInfo.Accrual)
		if err != nil {
			return fmt.Errorf("failed to updage balance current amount: %w", err)
		}

		return nil
	})

	return err
}
