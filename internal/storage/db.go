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

	CreateOrder(ctx context.Context, userID int, number string, status models.OrderStatus) (models.Order, error)
	FindOrderByNumber(ctx context.Context, number string) (models.Order, error)

	BeginTranscaction(ctx context.Context) (pgx.Tx, error)
}

type DBStorage struct {
	pool *pgxpool.Pool
}

type ErrUserNotUniq struct {
	User models.User
}

type ErrOrderNotUnique struct {
	order models.Order
}

type ErrUserNotFound struct {
	User models.User
}

type ErrOrderNotFound struct {
	Order models.Order
}

func (err ErrUserNotUniq) Error() string {
	return fmt.Sprintf("user with login \"%s\" already exists", err.User.Login)
}

func (err ErrOrderNotUnique) Error() string {
	return fmt.Sprintf("order with number \"%s\" already exists", err.order.Number)
}

func (err ErrUserNotFound) Error() string {
	return fmt.Sprintf("user with login \"%s\" not found", err.User.Login)
}

func (err ErrOrderNotFound) Error() string {
	return fmt.Sprintf("order with number \"%s\" not found", err.Order.Number)
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
			"userID":   userID,
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
