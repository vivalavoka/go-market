package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/handlers"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
)

type Server struct {
	config   config.Config
	storage  *storage.Storage
	handlers *handlers.Handlers
}

func New(cfg config.Config, storage *storage.Storage) *Server {
	return &Server{
		config:  cfg,
		storage: storage,
	}
}

func (s *Server) Start() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s.handlers = handlers.New(s.config, s.storage)
	s.handlers.SetRoutes(r)

	http.ListenAndServe(s.config.Address, r)
}
