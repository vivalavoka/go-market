package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/vivalavoka/go-market/cmd/gophermart/http/middlewares"
)

func (h *Handlers) SetRoutes(r chi.Router) chi.Router {
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)

	r.Route("/api/user", func(ri chi.Router) {
		ri.Use(middlewares.CheckToken)
		ri.Post("/orders", h.CreateOrder)
		ri.Get("/orders", h.OrderList)
		ri.Get("/balance", h.GetBalance)
		ri.Post("/balance/withdraw", h.Withdraw)
		ri.Get("/withdrawals", h.Withdrawals)
	})

	r.Get("/api/orders/{number}", h.EchoAccrualHandler)
	return r
}
