package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/handlers"
	"github.com/vivalavoka/go-market/cmd/gophermart/http/middlewares"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
)

type Server struct {
	storage  *storage.Storage
	handlers *handlers.Handlers
}

func New(storage *storage.Storage) *Server {
	return &Server{
		storage: storage,
	}
}

func (s *Server) Start(cfg config.Config) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CompressHandle)
	r.Use(middlewares.DecompressHandle)

	s.handlers = handlers.New(cfg, s.storage)
	s.handlers.SetRoutes(r)

	http.ListenAndServe(cfg.Address, r)
}
