package data

import (
	"context"
	auth2 "delivery_service/auth_proto"
	"delivery_service/internal/validator"

	gosocketio "github.com/graarh/golang-socketio"
	"google.golang.org/grpc"
	"time"
)

type User struct {
	ID       int64    `json:"id"`
	UserName string   `json:"userName"`
	Email    string   `json:"email"`
	Type     string   `json:"type"`
	Location Location `json:"location,omitempty"`
	Channel  *gosocketio.Channel
	Version  int `json:"-"`
}

type DeliveryMessage struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	OrderId  int64    `json:"roomId"`
	Location Location `json:"location"`
}
type Location struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}

func ValidateLocations(v *validator.Validator, destLat, destLong, originLat, originLong float64) {
	v.Check(originLat >= 0 && originLong >= 0 && destLat >= 0 && destLong >= 0, "locations", "is not valid")
}

type UserModel struct {
	Conn *grpc.ClientConn
}

func (m *UserModel) GetUserForToken(token string) (*User, error) {

	c := auth2.NewAuthServiceClient(m.Conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.Authenticate(ctx, &auth2.AuthRequest{
		TokenEntry: &auth2.Token{
			Token: token,
		},
	})
	if err != nil {
		return nil, err
	}
	return &User{
		ID:       res.User.Id,
		UserName: res.User.UserName,
		Email:    res.User.Email,
		Type:     res.User.Type,
		Location: Location{},
		Channel:  nil,
	}, nil
}
