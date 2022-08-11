package middlewares

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"github.com/vivalavoka/go-market/cmd/gophermart/users"
)

const userIDKey = "user_id"
const loginKey = "login"

func CheckToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claim, err := r.Cookie("token")

		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token := claim.Value

		userClaim := &users.UserClaims{}
		result, err := jwt.ParseWithClaims(token, userClaim, func(token *jwt.Token) (interface{}, error) {
			return []byte(""), nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !result.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, userIDKey, strconv.FormatInt(int64(userClaim.ID), 10))
		ctx = context.WithValue(ctx, loginKey, userClaim.Login)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserClaim(ctx context.Context) *users.UserClaims {
	if ctx == nil {
		return nil
	}

	userClaim := users.UserClaims{}

	if UserIDStr, ok := ctx.Value(userIDKey).(string); ok {
		if UserID, err := strconv.ParseInt(UserIDStr, 10, 64); err == nil {
			userClaim.ID = users.PostgresPK(UserID)
		}
	}

	if login, ok := ctx.Value(loginKey).(string); ok {
		userClaim.Login = login
	}

	return &userClaim
}
