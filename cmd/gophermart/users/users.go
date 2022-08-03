package users

type PostgresPK int64

type User struct {
	ID       PostgresPK `json:"user_id" db:"user_id"`
	Login    string     `json:"login" db:"login"`
	Password string     `json:"password" db:"password"`
}

type UserBalance struct {
	ID    PostgresPK `json:"user_id" db:"user_id"`
	Value PostgresPK `json:"value" db:"value"`
}

type UserOrder struct {
	ID      PostgresPK `json:"user_id" db:"user_id"`
	OrderID PostgresPK `json:"value" db:"value"`
}
