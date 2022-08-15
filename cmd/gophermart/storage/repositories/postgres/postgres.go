package postgresdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

var ErrNotFound = errors.New("not found")

type PostgresDB struct {
	config         config.Config
	connection     *sqlx.DB
	createUserStmt *sql.Stmt
}

func New(cfg config.Config) (*PostgresDB, error) {
	conn, err := sqlx.Open("postgres", cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	postgres := PostgresDB{config: cfg, connection: conn}

	err = postgres.migration()

	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	if postgres.createUserStmt, err = postgres.connection.Prepare(
		`INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id;`,
	); err != nil {
		return nil, err
	}

	return &postgres, nil
}

func (r *PostgresDB) migration() error {
	var err error

	err = r.createUserTable()
	if err != nil {
		return err
	}

	err = r.createUserOrderTable()
	if err != nil {
		return err
	}

	err = r.createWithdrawTable()
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresDB) createUserTable() error {
	rows, err := r.connection.Query(`
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL,
			login VARCHAR UNIQUE,
			password VARCHAR,
			current NUMERIC(10, 2) DEFAULT 0,
			withdrawn NUMERIC(10, 2) DEFAULT 0
		);`)
	if rows.Err() != nil {
		return rows.Err()
	}
	return err
}

func (r *PostgresDB) createUserOrderTable() error {
	rows, err := r.connection.Query(`
		CREATE TABLE IF NOT EXISTS user_orders (
			user_id VARCHAR,
			number VARCHAR UNIQUE,
			accrual NUMERIC(10, 2) DEFAULT 0,
			status VARCHAR,
			uploaded_at TIMESTAMPTZ DEFAULT now()
		);`)
	if rows.Err() != nil {
		return rows.Err()
	}
	return err
}

func (r *PostgresDB) createWithdrawTable() error {
	rows, err := r.connection.Query(`
		CREATE TABLE IF NOT EXISTS user_withdrawals (
			user_id VARCHAR,
			number VARCHAR UNIQUE,
			sum NUMERIC(10, 2) DEFAULT 0,
			processed_at TIMESTAMPTZ DEFAULT now()
		);`)
	if rows.Err() != nil {
		return rows.Err()
	}
	return err
}

func (r *PostgresDB) Close() {
	r.connection.Close()
}

func (r *PostgresDB) CheckConnection() bool {
	err := r.connection.Ping()
	return err == nil
}

func (r *PostgresDB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.connection.BeginTx(ctx, nil)
}

func (r *PostgresDB) CreateUser(user *users.User) error {
	_, err := r.createUserStmt.Exec(user.Login, user.Password)

	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresDB) GetUserByLogin(login string) (*users.User, error) {
	var data []users.User
	err := r.connection.Select(&data, `SELECT user_id, login, password FROM users WHERE login = $1 LIMIT 1;`, login)

	if len(data) == 0 {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &data[0], nil
}

func (r *PostgresDB) GetUserBalance(UserID users.PostgresPK) (*users.User, error) {
	var data []users.User
	err := r.connection.Select(&data, `SELECT current, withdrawn FROM users WHERE user_id = $1 LIMIT 1;`, UserID)

	if len(data) == 0 {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &data[0], nil
}

func (r *PostgresDB) IncreaseUserBalance(tx *sql.Tx, UserID users.PostgresPK, value float32) error {
	_, err := tx.ExecContext(context.Background(), "UPDATE users SET current = current + $1 WHERE user_id = $2;", value, UserID)

	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresDB) DecreaseUserBalance(tx *sql.Tx, UserID users.PostgresPK, value float32) error {
	_, err := tx.ExecContext(context.Background(), "UPDATE users SET current = current - $1, withdrawn = withdrawn + $1 WHERE user_id = $2;", value, UserID)

	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresDB) GetOrder(orderID string) (*users.UserOrder, error) {
	var data []users.UserOrder
	err := r.connection.Select(&data, `SELECT user_id, number, status, accrual, uploaded_at FROM user_orders WHERE number = $1 LIMIT 1;`, orderID)

	if len(data) == 0 {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &data[0], nil
}

func (r *PostgresDB) UpsertOrder(tx *sql.Tx, userOrder *users.UserOrder) error {
	_, err := tx.ExecContext(context.Background(), "INSERT INTO user_orders (user_id, number, status, accrual) VALUES ($1, $2, $3, $4) ON CONFLICT (number) DO UPDATE SET status = $3, accrual = $4;",
		userOrder.UserID, userOrder.Number, userOrder.Status, userOrder.Accrual)

	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresDB) GetOrderList(UserID users.PostgresPK) ([]users.UserOrder, error) {
	var data []users.UserOrder
	err := r.connection.Select(&data, `SELECT number, status, accrual, uploaded_at FROM user_orders WHERE user_id = $1 ORDER BY uploaded_at ASC;`, UserID)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *PostgresDB) GetOrdersByStatus(status string) ([]users.UserOrder, error) {
	var data []users.UserOrder
	err := r.connection.Select(&data, `SELECT user_id, number, accrual, status FROM user_orders WHERE status = $1 ORDER BY uploaded_at ASC;`, status)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *PostgresDB) CreateWithdraw(tx *sql.Tx, withdraw users.UserWithdraw) error {
	_, err := tx.ExecContext(context.Background(), "INSERT INTO user_withdrawals (user_id, number, sum) VALUES ($1, $2, $3);",
		withdraw.UserID, withdraw.Number, withdraw.Sum)

	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresDB) GetWithdrawals(UserID users.PostgresPK) ([]users.UserWithdraw, error) {
	var data []users.UserWithdraw
	err := r.connection.Select(&data, `SELECT number, sum, processed_at FROM user_withdrawals WHERE user_id = $1 ORDER BY processed_at ASC;`, UserID)

	if err != nil {
		return nil, err
	}

	return data, nil
}
