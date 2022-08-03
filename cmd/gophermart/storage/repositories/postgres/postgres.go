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
	config     config.Config
	connection *sqlx.DB
	createStmt *sql.Stmt
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

	postgres.createStmt, err = postgres.connection.Prepare(
		`INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id;`,
	)

	if err != nil {
		return nil, err
	}

	return &postgres, nil
}

func (r *PostgresDB) migration() error {
	rows, err := r.connection.Query(`
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL,
			login VARCHAR UNIQUE,
			password VARCHAR
		);`,
	)
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
	_, err := r.createStmt.Exec(user.Login, user.Password)

	if err != nil {
		pgError := err.(*pg.Error)
		return fmt.Sprint(pgError.Code)
	}
	return ""
}

func (r *PostgresDB) GetUserByLogin(login string) (users.User, error) {
	var data users.User
	err := r.connection.Get(&data, `SELECT user_id, login, password FROM users WHERE login = $1;`, login)

	log.Info(data)
	if err != nil {
		return users.User{}, err
	}

	return data, nil
}

func (r *PostgresDB) GetUserById(userId string) (users.User, error) {
	var data users.User
	err := r.connection.Get(&data, `SELECT user_id, login, password FROM users WHERE user_id = $1;`, userId)

	log.Info(data)
	if err != nil {
		return users.User{}, err
	}

	return data, nil
}
