package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	CreateUser(ctx context.Context, login, encryptedPassword string) (models.User, error)
	FindUserByLogin(ctx context.Context, login string) (models.User, error)
	UserOrders(ctx context.Context, userID int) ([]models.Order, error)

	CreateOrder(ctx context.Context, userID int, number string, status models.OrderStatus) (models.Order, error)
	FindOrderByNumber(ctx context.Context, number string) (models.Order, error)
	UpdateOrder(ctx context.Context, orderID int, status models.OrderStatus, accrual int) error

	CreateBalance(ctx context.Context, userID, currentAmount int) (models.Balance, error)
	UpdateBalanceCurrentAmount(ctx context.Context, balanceID, amount int) error
	FindBalanceByUserID(ctx context.Context, userID int) (models.Balance, error)

	UserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)

	BeginTranscaction(ctx context.Context) (pgx.Tx, error)
}

type DBStorage struct {
	pool *pgxpool.Pool
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	pool, err := pgxpool.New(context.TODO(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}

	return &DBStorage{
		pool: pool,
	}, nil
}

func (db *DBStorage) CreateUser(ctx context.Context, login, encryptedPassword string) (models.User, error) {
	row := db.pool.QueryRow(
		ctx,
		`INSERT INTO "users" ("login", "encrypted_password") VALUES (@login, @encryptedPassword) RETURNING "id"`,
		pgx.NamedArgs{"login": login, "encryptedPassword": encryptedPassword},
	)
	var userID int
	user := models.User{Login: login, EncryptedPassword: encryptedPassword}
	err := row.Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return user, ErrUserNotUniq{User: user}
		}
		return user, fmt.Errorf("failed to create user %w", err)
	}
	user.ID = userID

	return user, nil
}

func (db *DBStorage) FindUserByLogin(ctx context.Context, login string) (models.User, error) {
	row := db.pool.QueryRow(
		ctx,
		`SELECT "id", "encrypted_password"
		 FROM "users"
		 WHERE "login" = @login`,
		pgx.NamedArgs{"login": login},
	)
	user := models.User{Login: login}
	var id int
	var encryptedPassword string
	err := row.Scan(&id, &encryptedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, ErrUserNotFound{User: user}
		}
		return user, fmt.Errorf("failed to find user: %w", err)
	}

	user.ID = id
	user.EncryptedPassword = encryptedPassword

	return user, nil
}

func (db *DBStorage) UserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	rows, err := db.pool.Query(
		ctx,
		`SELECT "id", "user_id", "number", "status", "accrual", "created_at"
		 FROM "orders"
		 WHERE "user_id" = @userID
		 ORDER BY "created_at" DESC`,
		pgx.NamedArgs{"userID": userID},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	result, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Order, error) {
		var (
			id        int
			userID    int
			number    string
			status    models.OrderStatus
			accrual   int
			createdAt time.Time
		)
		err := row.Scan(&id, &userID, &number, &status, &accrual, &createdAt)

		return models.Order{
			ID:        id,
			UserID:    userID,
			Number:    number,
			Status:    status,
			Accrual:   accrual,
			CreatedAt: createdAt,
		}, err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	return result, nil
}

func (db *DBStorage) CreateOrder(
	ctx context.Context,
	userID int,
	number string,
	status models.OrderStatus) (models.Order, error) {

	currentTime := time.Now()
	row := db.pool.QueryRow(
		ctx,
		`INSERT INTO "orders" ("user_id", "number", "status", "created_at")
		 VALUES (@userID, @number, @status, @createdAt) RETURNING "id"`,
		pgx.NamedArgs{
			"userID":    userID,
			"number":    number,
			"status":    status,
			"createdAt": currentTime,
		},
	)
	var orderID int
	order := models.Order{
		UserID:    userID,
		Number:    number,
		Status:    status,
		CreatedAt: currentTime,
	}
	err := row.Scan(&orderID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return order, ErrOrderNotUnique{order: order}
		}
		return order, fmt.Errorf("failed to create order: %w", err)
	}
	order.ID = orderID

	return order, nil
}

func (db *DBStorage) FindOrderByNumber(ctx context.Context, number string) (models.Order, error) {
	row := db.pool.QueryRow(
		ctx,
		`SELECT "id", "user_id", "status", "accrual", "created_at"
		 FROM "orders"
		 WHERE "number" = @number`,
		pgx.NamedArgs{"number": number},
	)
	order := models.Order{Number: number}
	var id, userID, accrual int
	var status models.OrderStatus
	var createdAt time.Time
	err := row.Scan(&id, &userID, &status, &accrual, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return order, ErrOrderNotFound{Order: order}
		}
		return order, fmt.Errorf("failed to find order: %w", err)
	}

	order.UserID = userID
	order.Status = status
	order.Accrual = accrual
	order.CreatedAt = createdAt

	return order, nil
}

func (db *DBStorage) CreateBalance(ctx context.Context, userID, currentAmount int) (models.Balance, error) {
	row := db.pool.QueryRow(
		ctx,
		`INSERT INTO "balances" ("user_id", "current_amount")
		 VALUES (@userID, @currentAmount) RETURNING "id"`,
		pgx.NamedArgs{"userID": userID, "currentAmount": currentAmount},
	)
	var balanceID int
	balance := models.Balance{
		UserID:        userID,
		CurrentAmount: currentAmount,
	}
	err := row.Scan(&balanceID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return balance, ErrBalanceNotUnique{Balance: balance}
		}

		return balance, fmt.Errorf("failed to create balance: %w", err)
	}
	balance.ID = balanceID

	return balance, nil
}

func (db *DBStorage) UpdateBalanceCurrentAmount(ctx context.Context, balanceID, amount int) error {
	_, err := db.pool.Exec(
		ctx,
		`UPDATE "balances" SET "current_amount" = @currentAmount WHERE "id" = @balanceID`,
		pgx.NamedArgs{"currentAmount": amount, "balanceID": balanceID},
	)
	if err != nil {
		return fmt.Errorf("failed to update amount for balance id=%d: %w", balanceID, err)
	}

	return nil
}

func (db *DBStorage) FindBalanceByUserID(ctx context.Context, userID int) (models.Balance, error) {
	row := db.pool.QueryRow(
		ctx,
		`SELECT "id", "current_amount", "withdrawn_amount" FROM "balances" WHERE "user_id" = @userID`,
		pgx.NamedArgs{"userID": userID},
	)
	balance := models.Balance{UserID: userID}
	var id, currentAmount, withdrawnAmount int
	err := row.Scan(&id, &currentAmount, &withdrawnAmount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return balance, ErrBalanceNotFound{Balance: balance}
		}
		return balance, fmt.Errorf("failed to find balance: %w", err)
	}
	balance.ID = id
	balance.CurrentAmount = currentAmount
	balance.WithdrawnAmount = withdrawnAmount

	return balance, nil
}

func (db *DBStorage) UpdateOrder(ctx context.Context, orderID int, status models.OrderStatus, accrual int) error {
	_, err := db.pool.Exec(
		ctx,
		`UPDATE "orders" SET "status" = @status, "accrual" = @accrual WHERE "id" = @orderID`,
		pgx.NamedArgs{"status": status, "accrual": accrual, "orderID": orderID},
	)
	if err != nil {
		return fmt.Errorf("failed to update status for order id=%d: %w", orderID, err)
	}

	return nil
}

func (db *DBStorage) UserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	rows, err := db.pool.Query(
		ctx,
		`SELECT "id", "order_number", "user_id", "sum", "processed_at"
		 FROM "withdrawals"
		 WHERE "user_id" = @userID
		 ORDER BY "processed_at"`,
		pgx.NamedArgs{"userID": userID},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch withdrawals: %w", err)
	}

	result, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Withdrawal, error) {
		var (
			id          int
			orderNumber string
			userID      int
			sum         int
			processedAt time.Time
		)
		err := row.Scan(&id, &orderNumber, &userID, &sum, &processedAt)

		return models.Withdrawal{
			ID:          id,
			OrderNumber: orderNumber,
			UserID:      userID,
			Sum:         sum,
			ProcessedAt: processedAt,
		}, err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch withdrawals: %w", err)
	}

	return result, nil
}

func (db *DBStorage) BeginTranscaction(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

//go:embed db/migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "db/migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	return nil
}
