package agent

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vivalavoka/go-market/cmd/gophermart/agent/client"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

const SyncInterval = time.Duration(500 * time.Millisecond)

type ClientInterface interface {
	GetAccrual(string) (*client.AccrualResponse, error)
}

type Agent struct {
	config  config.Config
	client  ClientInterface
	storage *storage.Storage
}

func New(cfg config.Config, stg *storage.Storage, clnt ClientInterface) *Agent {

	return &Agent{
		config:  cfg,
		client:  clnt,
		storage: stg,
	}
}

func (a *Agent) Start() {
	syncTicker := time.NewTicker(SyncInterval)
	defer syncTicker.Stop()

	for {
		<-syncTicker.C

		newOrders, err := a.storage.Repo.GetOrdersByStatus(users.New)
		if err != nil {
			log.Error(err)
		}

		if len(newOrders) > 0 {
			log.Info("New orders")
			a.accrualOrders(newOrders)
		}

		processingOrders, err := a.storage.Repo.GetOrdersByStatus(users.Processing)
		if err != nil {
			log.Error(err)
		}

		if len(processingOrders) > 0 {
			log.Info("Processing orders")
			a.accrualOrders(processingOrders)
		}
	}
}

func (a *Agent) accrualOrders(orders []users.UserOrder) {
	for _, order := range orders {
		response, err := a.client.GetAccrual(order.Number)
		if err != nil {
			log.Error(err)
			continue
		}

		a.processOrderByAccrual(order, *response)
	}
}

func (a *Agent) processOrderByAccrual(order users.UserOrder, accrual client.AccrualResponse) {
	tx, txErr := a.storage.Repo.BeginTx(context.Background())
	if txErr != nil {
		log.Error(txErr)
		return
	}
	defer tx.Rollback()

	switch accrual.Status {
	case "REGISTERED":
		return
	case "PROCESSING":
		order.Status = users.Processing
	case "INVALID":
		order.Status = users.Invalid
	case "PROCESSED":
		{
			order.Status = users.Processed
			order.Accrual = float32(accrual.Accrual)
			err := a.storage.Repo.IncreaseUserBalance(tx, order.UserID, order.Accrual)
			if err != nil {
				log.Error(err)
				return
			}
		}
	}

	err := a.storage.Repo.UpsertOrder(tx, &order)
	if err != nil {
		log.Error(err)
	}

	if err = tx.Commit(); err != nil {
		log.Error(err)
		return
	}
}
