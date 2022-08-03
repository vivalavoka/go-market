package handlers

import "github.com/go-chi/chi/v5"

func (h *Handlers) SetRoutes(r chi.Router) chi.Router {
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)
	r.Post("/api/user/orders", h.ProcessOrder)
	r.Get("/api/user/orders", h.OrderList)
	r.Get("/api/user/balance", h.GetBalance)
	r.Post("/api/user/withdraw", h.Withdraw)
	r.Get("/api/user/withdrawals", h.Withdrawals)

	return r
}
