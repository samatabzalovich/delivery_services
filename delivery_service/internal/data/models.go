package data

import (
	"database/sql"
	"errors"
	"google.golang.org/grpc"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)
var db *sql.DB

func New(connection *grpc.ClientConn, dbPool *sql.DB) Models {

	return Models{
		Users: &UserModel{Conn: connection},
		Order: &OrderModel{DB: dbPool},
	}
}

type Models struct {
	Users interface {
		GetUserForToken(token string) (*User, error)
	}
	Order interface {
		CompleteOrder(id int64) error
		UpdateOrder(order *Order) error
		GetOrderForUser(user *User) (*Order, error)
		Insert(order *Order) error
	}
}
