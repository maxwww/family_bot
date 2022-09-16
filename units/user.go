package units

import (
	"context"
)

type User struct {
	ID            uint
	TelegramID    uint   `db:"telegram_id"`
	FirstName     string `db:"first_name"`
	LastName      string `db:"last_name"`
	UserName      string `db:"user_name"`
	Notifications bool   `db:"notifications"`
}

type UserPatch struct {
	FirstName     *string
	LastName      *string
	UserName      *string
	Notifications *bool
}

type UserFilter struct {
	TelegramID *uint

	Limit  int
	Offset int
}

type UserService interface {
	CreateUser(context.Context, *User) error

	UserByTelegramID(context.Context, uint) (*User, error)

	Users(context.Context, UserFilter) ([]*User, error)

	UpdateUser(context.Context, *User, UserPatch) error
}
