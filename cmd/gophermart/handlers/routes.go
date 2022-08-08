package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/vivalavoka/go-market/cmd/gophermart/http/middlewares"
)

func (h *Handlers) SetRoutes(r chi.Router) chi.Router {
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)

	r.Route("/api/user/orders", func(ri chi.Router) {
		ri.Use(middlewares.CheckToken)
		ri.Post("/", h.LinkOrder)
		ri.Get("/", h.OrderList)
	})

	r.Route("/api/user/balance", func(ri chi.Router) {
		ri.Use(middlewares.CheckToken)
		ri.Get("/", h.GetBalance)
	})

	r.Route("/api/user/withdraw", func(ri chi.Router) {
		ri.Use(middlewares.CheckToken)
		ri.Post("/", h.Withdraw)
	})

	r.Route("/api/user/withdrawals", func(ri chi.Router) {
		ri.Use(middlewares.CheckToken)
		ri.Get("/", h.Withdrawals)
	})

	return r
}
