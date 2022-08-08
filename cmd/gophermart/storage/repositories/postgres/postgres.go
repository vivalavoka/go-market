package postgresdb

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	pg "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

type PostgresDB struct {
	config            config.Config
	connection        *sqlx.DB
	createUserStmt    *sql.Stmt
	createBalanceStmt *sql.Stmt
	linkOrderStmt     *sql.Stmt
}

func New(cfg config.Config) (*PostgresDB, error) {
	conn, err := sqlx.Open("postgres", cfg.DatabaseUri)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	postgres := PostgresDB{config: cfg, connection: conn}

	err = postgres.migration()

	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	postgres.createUserStmt, err = postgres.connection.Prepare(
		`INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id;`,
	)

	postgres.linkOrderStmt, err = postgres.connection.Prepare(
		`INSERT INTO user_orders (user_id, order_id) VALUES ($1, $2);`,
	)

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

	return nil
}

func (r *PostgresDB) createUserTable() error {
	rows, err := r.connection.Query(`
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL,
			login VARCHAR UNIQUE,
			password VARCHAR,
			balance INTEGER DEFAULT 0
		);`)
	if rows.Err() != nil {
		return rows.Err()
	}
	return err
}

func (r *PostgresDB) createUserOrderTable() error {
	rows, err := r.connection.Query(`
		CREATE TABLE IF NOT EXISTS user_orders (
			user_id VARCHAR PRIMARY KEY,
			order_id VARCHAR UNIQUE
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

func (r *PostgresDB) GetUserBalance(userId users.PostgresPK) (*users.User, error) {
	var data []users.User
	err := r.connection.Select(&data, `SELECT balance FROM users WHERE user_id = $1 LIMIT 1;`, userId)

	log.Info(data)
	if len(data) == 0 {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &data[0], nil
}

func (r *PostgresDB) GetOrder(orderId users.PostgresPK) (*users.UserOrder, error) {
	var data []users.UserOrder
	err := r.connection.Select(&data, `SELECT user_id, order_id FROM user_orders WHERE order_id = $1 LIMIT 1;`, orderId)

	if len(data) == 0 {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &data[0], nil
}

func (r *PostgresDB) GetOrderList(userId users.PostgresPK) ([]users.UserOrder, error) {
	var data []users.UserOrder
	err := r.connection.Select(&data, `SELECT user_id, order_id FROM user_orders WHERE user_id = $1`, userId)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *PostgresDB) LinkOrder(userOrder *users.UserOrder) string {
	_, err := r.linkOrderStmt.Exec(userOrder.UserId, userOrder.OrderID)

	if err != nil {
		pgError := err.(*pg.Error)
		log.Error(pgError)
		return fmt.Sprint(pgError.Code)
	}
	return ""
}
