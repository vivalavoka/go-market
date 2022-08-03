package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

type Handlers struct {
	storage *storage.Storage
}

func New(cfg config.Config, storage *storage.Storage) *Handlers {
	return &Handlers{
		storage: storage,
	}
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var params *users.User

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	errCode := h.storage.Repo.CreateUser(params)

	if errCode != "" {
		if errCode == "23505" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var params *users.User

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.storage.Repo.GetUserByLogin(params.Login)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.Password != params.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (h *Handlers) OrderList(w http.ResponseWriter, r *http.Request) {
}

func (h *Handlers) ProcessOrder(w http.ResponseWriter, r *http.Request) {
}

func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
}

func (h *Handlers) Withdraw(w http.ResponseWriter, r *http.Request) {
}

func (h *Handlers) Withdrawals(w http.ResponseWriter, r *http.Request) {
}
