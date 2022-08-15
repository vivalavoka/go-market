package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/omeid/pgerror"
	"github.com/vivalavoka/go-market/cmd/gophermart/config"
	"github.com/vivalavoka/go-market/cmd/gophermart/http/middlewares"
	"github.com/vivalavoka/go-market/cmd/gophermart/storage"
	postgresdb "github.com/vivalavoka/go-market/cmd/gophermart/storage/repositories/postgres"
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

func (h *Handlers) EchoAccrualHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	number := chi.URLParam(r, "number")
	w.Write([]byte(fmt.Sprintf(`{"order": "%s","status": "PROCESSED","accrual": 120.87}`, number)))
}

func (h *Handlers) auth(w http.ResponseWriter, params *users.User) {
	user, err := h.storage.Repo.GetUserByLogin(params.Login)

	if err != nil {
		if errors.Is(err, postgresdb.ErrNotFound) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if user.Password != params.Password {
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

}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var params *users.User

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.Repo.CreateUser(params)

	if err != nil {
		if e := pgerror.UniqueViolation(err); e != nil {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.auth(w, params)

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

	h.auth(w, params)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (h *Handlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	param := buf.String()

	orderID, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !luhn.Valid(orderID) {
		http.Error(w, "Invalid order id format", http.StatusUnprocessableEntity)
		return
	}

	order, pgErr := h.storage.Repo.GetOrder(param)

	if pgErr != nil && !errors.Is(pgErr, postgresdb.ErrNotFound) {
		http.Error(w, pgErr.Error(), http.StatusInternalServerError)
		return
	}

	session := middlewares.GetUserClaim(r.Context())

	if order != nil {
		if order.UserID == session.ID {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("номер заказа уже был загружен этим пользователем"))
			return
		} else {
			http.Error(w, "номер заказа уже был загружен другим пользователем", http.StatusConflict)
			return
		}
	}
	tx, txErr := h.storage.Repo.BeginTx(r.Context())
	if txErr != nil {
		http.Error(w, txErr.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback()

	err = h.storage.Repo.UpsertOrder(tx, &users.UserOrder{UserID: session.ID, Number: param, Status: users.New})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) OrderList(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetUserClaim(r.Context())

	orders, pgErr := h.storage.Repo.GetOrderList(session.ID)
	if pgErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(pgErr.Error()))
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetUserClaim(r.Context())

	balance, pgErr := h.storage.Repo.GetUserBalance(session.ID)
	if pgErr != nil {
		http.Error(w, pgErr.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handlers) Withdraw(w http.ResponseWriter, r *http.Request) {
	var params *users.UserWithdraw
	session := middlewares.GetUserClaim(r.Context())

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.UserID = session.ID

	orderID, err := strconv.ParseInt(params.Number, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !luhn.Valid(orderID) {
		http.Error(w, "Invalid order id format", http.StatusUnprocessableEntity)
		return
	}

	user, err := h.storage.Repo.GetUserBalance(session.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.Current < params.Sum {
		http.Error(w, "Not enough funds", http.StatusPaymentRequired)
		return
	}

	tx, txErr := h.storage.Repo.BeginTx(r.Context())
	if txErr != nil {
		http.Error(w, txErr.Error(), http.StatusInternalServerError)
	}
	defer tx.Rollback()

	err = h.storage.Repo.DecreaseUserBalance(tx, session.ID, params.Sum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.storage.Repo.CreateWithdraw(tx, *params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (h *Handlers) Withdrawals(w http.ResponseWriter, r *http.Request) {
	session := middlewares.GetUserClaim(r.Context())

	withdrawals, pgErr := h.storage.Repo.GetWithdrawals(session.ID)
	if pgErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(pgErr.Error()))
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(withdrawals)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
