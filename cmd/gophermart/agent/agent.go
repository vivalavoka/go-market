package agent

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vivalavoka/go-market/cmd/gophermart/agent/client"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

const SYNC_INTERVAL = time.Duration(2 * time.Second)

type Agent struct {
	config  config.Config
	client  *client.Client
	storage *storage.Storage
}

type AccrualResponse struct {
	Order   string
	Status  string
	Accrual int
}

func New(cfg config.Config, stg *storage.Storage) *Agent {
	client := client.New(cfg.AccrualSystemAddress)

	return &Agent{
		config:  cfg,
		client:  client,
		storage: stg,
	}
}

func (a *Agent) Start() {
	syncTicker := time.NewTicker(SYNC_INTERVAL)
	defer syncTicker.Stop()

	for {
		<-syncTicker.C

		log.Info("Start sync")
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
			order.Accrual = int16(accrual.Accrual)
			a.storage.Repo.UpdateUserBalance(order.UserId, order.Accrual)
		}
	}

	a.storage.Repo.UpsertOrder(&order)
}
