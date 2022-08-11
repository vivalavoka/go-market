package users

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type PostgresPK int64

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type UserClaims struct {
	ID    PostgresPK `json:"user_id"`
	Login string     `json:"login"`
	jwt.StandardClaims
}

type User struct {
	ID        PostgresPK `json:"user_id,omitempty" db:"user_id"`
	Login     string     `json:"login,omitempty" db:"login"`
	Password  string     `json:"password,omitempty" db:"password"`
	Current   int16      `json:"current" db:"current"`
	Withdrawn int16      `json:"withdrawn" db:"withdrawn"`
}

const (
	New        string = "NEW"
	Processing        = "PROCESSING"
	Invalid           = "INVALID"
	Processed         = "PROCESSED"
)

type UserOrder struct {
	UserId     PostgresPK `json:"user_id,omitempty" db:"user_id"`
	Number     string     `json:"number" db:"number"`
	Accrual    int16      `json:"accrual,omitempty" db:"accrual"`
	Status     string     `json:"status" db:"status"`
	UploadedAt time.Time  `json:"uploaded_at" db:"uploaded_at"`
}

type UserWithdraw struct {
	UserId      PostgresPK `json:"user_id,omitempty" db:"user_id"`
	Number      string     `json:"order" db:"number"`
	Sum         int16      `json:"sum,omitempty" db:"sum"`
	ProcessedAt time.Time  `json:"processed_at" db:"processed_at"`
}
