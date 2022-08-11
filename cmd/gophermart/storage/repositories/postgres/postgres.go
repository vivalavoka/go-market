package postgresdb

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	pg "github.com/lib/pq"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

type PostgresDB struct {
	config                 config.Config
	connection             *sqlx.DB
	createUserStmt         *sql.Stmt
	createUserWithdrawStmt *sql.Stmt
	upsertOrderStmt        *sql.Stmt
	increaseBalanceStmt    *sql.Stmt
	decreaseBalanceStmt    *sql.Stmt
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

	if postgres.createUserWithdrawStmt, err = postgres.connection.Prepare(
		`INSERT INTO user_withdrawals (user_id, number, sum) VALUES ($1, $2, $3);`,
	); err != nil {
		return nil, err
	}

	if postgres.upsertOrderStmt, err = postgres.connection.Prepare(
		`INSERT INTO user_orders (user_id, number, status, accrual)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (number) DO UPDATE SET status = $3, accrual = $4;`,
	); err != nil {
		return nil, err
	}

	if postgres.increaseBalanceStmt, err = postgres.connection.Prepare(
		`UPDATE users SET current = current + $1 WHERE user_id = $2;`,
	); err != nil {
		return nil, err
	}

	if postgres.decreaseBalanceStmt, err = postgres.connection.Prepare(
		`UPDATE users SET current = current - $1, withdrawn = withdrawn + $1 WHERE user_id = $2;`,
	); err != nil {
		return nil, err
	}

	if err != nil {
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
			current INTEGER DEFAULT 0,
			withdrawn INTEGER DEFAULT 0
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
			accrual INTEGER DEFAULT 0,
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
			sum INTEGER DEFAULT 0,
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

func (r *PostgresDB) CreateUser(user *users.User) string {
	_, err := r.createUserStmt.Exec(user.Login, user.Password)

	if err != nil {
		pgError := err.(*pg.Error)
		return fmt.Sprint(pgError.Code)
	}
	return ""
}

func (r *PostgresDB) GetUserByLogin(login string) (*users.User, error) {
	var data []users.User
	err := r.connection.Select(&data, `SELECT user_id, login, password FROM users WHERE login = $1 LIMIT 1;`, login)

	if len(data) == 0 {
		return nil, nil
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
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &data[0], nil
}

func (r *PostgresDB) IncreaseUserBalance(UserID users.PostgresPK, value int16) string {
	_, err := r.increaseBalanceStmt.Exec(value, UserID)

	if err != nil {
		pgError := err.(*pg.Error)
		return fmt.Sprint(pgError.Code)
	}
	return ""
}

func (r *PostgresDB) DecreaseUserBalance(UserID users.PostgresPK, value int16) string {
	_, err := r.decreaseBalanceStmt.Exec(value, UserID)

	if err != nil {
		pgError := err.(*pg.Error)
		return fmt.Sprint(pgError.Code)
	}
	return ""
}

func (r *PostgresDB) GetOrder(orderID string) (*users.UserOrder, error) {
	var data []users.UserOrder
	err := r.connection.Select(&data, `SELECT user_id, number, status, accrual, uploaded_at FROM user_orders WHERE number = $1 LIMIT 1;`, orderID)

	if len(data) == 0 {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &data[0], nil
}

func (r *PostgresDB) UpsertOrder(userOrder *users.UserOrder) string {
	_, err := r.upsertOrderStmt.Exec(userOrder.UserID, userOrder.Number, userOrder.Status, userOrder.Accrual)

	if err != nil {
		pgError := err.(*pg.Error)
		return fmt.Sprint(pgError.Code)
	}
	return ""
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

func (r *PostgresDB) CreateWithdraw(withdraw users.UserWithdraw) string {
	_, err := r.createUserWithdrawStmt.Exec(withdraw.UserID, withdraw.Number, withdraw.Sum)

	if err != nil {
		pgError := err.(*pg.Error)
		return fmt.Sprint(pgError.Code)
	}
	return ""
}

func (r *PostgresDB) GetWithdrawals(UserID users.PostgresPK) ([]users.UserWithdraw, error) {
	var data []users.UserWithdraw
	err := r.connection.Select(&data, `SELECT number, sum, processed_at FROM user_withdrawals WHERE user_id = $1 ORDER BY processed_at ASC;`, UserID)

	if err != nil {
		return nil, err
	}

	return data, nil
}
