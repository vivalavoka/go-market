package users

import "github.com/dgrijalva/jwt-go"

type PostgresPK int64

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type UserClaims struct {
	ID    PostgresPK `json:"user_id"`
	Login string     `json:"login"`
	jwt.StandardClaims
}

type User struct {
	ID       PostgresPK `json:"user_id,omitempty" db:"user_id"`
	Login    string     `json:"login,omitempty" db:"login"`
	Password string     `json:"password,omitempty" db:"password"`
	Balance  int16      `json:"balance" db:"balance"`
}

type UserBalance struct {
	UserId PostgresPK `json:"user_id" db:"user_id"`
	Value  PostgresPK `json:"value" db:"value"`
}

type UserOrder struct {
	UserId  PostgresPK `json:"user_id" db:"user_id"`
	OrderID PostgresPK `json:"order_id" db:"order_id"`
}
