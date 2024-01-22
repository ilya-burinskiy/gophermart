package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
	"github.com/jackc/pgx/v5"
)

var ErrNotEnoughAmount = errors.New("not enough amount on balance")

type WithdrawalCreator interface {
	Call(ctx context.Context, userID int, orderNumber string, sum int) (models.Withdrawal, error)
}

func NewWithdrawalCreator(store storage.Storage) WithdrawalCreator {
	return withdrawalCreator{
		store: store,
	}
}

type withdrawalCreator struct {
	store storage.Storage
}

func (srv withdrawalCreator) Call(
	ctx context.Context,
	userID int,
	orderNumber string,
	sum int) (models.Withdrawal, error) {

	var withdrawal models.Withdrawal
	err := srv.store.WithinTranscaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		withdrawal, err = srv.store.CreateWithdrawalTx(ctx, tx, userID, orderNumber, sum)
		if err != nil {
			return err
		}

		balance, err := srv.store.FindBalanceByUserIDTx(ctx, tx, userID)
		if err != nil {
			var notFoundErr storage.ErrBalanceNotFound
			if errors.As(err, &notFoundErr) {
				balance, err = srv.store.CreateBalanceTx(ctx, tx, userID, 0)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("an unexpected error occured while trying to find balance: %w", err)
			}
		}

		if balance.CurrentAmount >= sum {
			err = srv.store.UpdateBalanceWithdrawnAmountTx(ctx, tx, balance.ID, balance.WithdrawnAmount+sum)
			if err != nil {
				return err
			}
			err = srv.store.UpdateBalanceCurrentAmountTx(ctx, tx, balance.ID, balance.CurrentAmount-sum)
			if err != nil {
				return err
			}

			return nil
		}

		return ErrNotEnoughAmount
	})

	return withdrawal, err
}
