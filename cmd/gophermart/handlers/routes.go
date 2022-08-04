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

	r.Get("/api/user/balance", h.GetBalance)
	r.Post("/api/user/withdraw", h.Withdraw)
	r.Get("/api/user/withdrawals", h.Withdrawals)

	return r
}
