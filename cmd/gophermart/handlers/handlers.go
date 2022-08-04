package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/http/middlewares"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
	"github.com/vivalavoka/go-market/internal/luhn"
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

	if user == nil || user.Password != params.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &users.UserClaims{
		ID:    user.ID,
		Login: user.Login,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(""))
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (h *Handlers) LinkOrder(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	param := buf.String()

	orderId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !luhn.Valid(orderId) {
		http.Error(w, "Invalid order id format", http.StatusUnprocessableEntity)
		return
	}

	order, pgErr := h.storage.Repo.GetOrder(users.PostgresPK(orderId))

	if pgErr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session := middlewares.GetUserClaim(r.Context())

	if order != nil {
		if order.UserId == session.ID {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("номер заказа уже был загружен этим пользователем"))
			return
		} else {
			http.Error(w, "номер заказа уже был загружен другим пользователем", http.StatusConflict)
			return
		}
	}

	h.storage.Repo.LinkOrder(&users.UserOrder{UserId: session.ID, OrderID: users.PostgresPK(orderId)})

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) OrderList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[]"))
}

func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
}

func (h *Handlers) Withdraw(w http.ResponseWriter, r *http.Request) {
}

func (h *Handlers) Withdrawals(w http.ResponseWriter, r *http.Request) {
}
