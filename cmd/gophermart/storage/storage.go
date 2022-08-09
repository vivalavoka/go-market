package storage

import (
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	postgresdb "github.com/vivalavoka/go-market/cmd/gophermart/storage/repositories/postgres"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

type MetricsRepoInterface interface {
	Close()
	CheckConnection() bool
	CreateUser(*users.User) string
	GetUserByLogin(string) (*users.User, error)
	GetUserBalance(users.PostgresPK) (*users.User, error)
	GetOrder(users.PostgresPK) (*users.UserOrder, error)
	UpsertOrder(*users.UserOrder) string
	GetOrderList(users.PostgresPK) ([]users.UserOrder, error)
	GetOrdersByStatus(status string) ([]users.UserOrder, error)
}

type Storage struct {
	Repo MetricsRepoInterface
}

func New(config config.Config) (*Storage, error) {
	var repo MetricsRepoInterface
	var err error

	repo, err = postgresdb.New(config)

	if err != nil {
		return nil, err
	}

	storage := &Storage{
		Repo: repo,
	}

	return storage, nil
}

func (s *Storage) Close() {
	s.Repo.Close()
}
