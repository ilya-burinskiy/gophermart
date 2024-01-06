package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ilya-burinskiy/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	CreateUser(ctx context.Context, login, encryptedPassword string) (models.User, error)
}

type DBStorage struct {
	pool *pgxpool.Pool
}

type ErrUserNotUniq struct {
	User models.User
}

func (err ErrUserNotUniq) Error() string {
	return fmt.Sprintf("user with login \"%s\" already exists", err.User.Login)
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
