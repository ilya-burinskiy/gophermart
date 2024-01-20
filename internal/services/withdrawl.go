package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/ilya-burinskiy/gophermart/internal/storage"
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

	tx, err := srv.store.BeginTranscaction(ctx)
	if err != nil {
		return models.Withdrawal{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	withdrawal, err := srv.store.CreateWithdrawal(ctx, userID, orderNumber, sum)
	if err != nil {
		return withdrawal, err
	}

	balance, err := srv.store.FindBalanceByUserID(ctx, userID)
	if err != nil {
		var notFoundErr storage.ErrBalanceNotFound
		if errors.As(err, &notFoundErr) {
			balance, err = srv.store.CreateBalance(ctx, userID, 0)
			if err != nil {
				return withdrawal, err
			}
		} else {
			return withdrawal, fmt.Errorf("an unexpected error occured while trying to find balance: %w", err)
		}
	}

	if balance.CurrentAmount >= sum {
		err = srv.store.UpdateBalanceWithdrawnAmount(ctx, balance.ID, balance.WithdrawnAmount+sum)
		if err != nil {
			return withdrawal, err
		}
		err = srv.store.UpdateBalanceCurrentAmount(ctx, balance.ID, balance.CurrentAmount-sum)
		if err != nil {
			return withdrawal, err
		}

		if err = tx.Commit(ctx); err != nil {
			return withdrawal, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return withdrawal, nil
	}

	return withdrawal, ErrNotEnoughAmount
}
