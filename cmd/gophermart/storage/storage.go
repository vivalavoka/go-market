package storage

import (
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	postgresdb "github.com/vivalavoka/go-market/cmd/gophermart/storage/repositories/postgres"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

type MarketRepoInterface interface {
	Close()
	CheckConnection() bool
	CreateUser(*users.User) error
	GetUserByLogin(string) (*users.User, error)
	GetUserBalance(users.PostgresPK) (*users.User, error)
	IncreaseUserBalance(users.PostgresPK, float32) error
	DecreaseUserBalance(users.PostgresPK, float32) error
	GetOrder(string) (*users.UserOrder, error)
	UpsertOrder(*users.UserOrder) error
	GetOrderList(users.PostgresPK) ([]users.UserOrder, error)
	GetOrdersByStatus(status string) ([]users.UserOrder, error)
	CreateWithdraw(users.UserWithdraw) error
	GetWithdrawals(users.PostgresPK) ([]users.UserWithdraw, error)
}

type Storage struct {
	Repo MarketRepoInterface
}

func New(config config.Config) (*Storage, error) {
	var repo MarketRepoInterface
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
