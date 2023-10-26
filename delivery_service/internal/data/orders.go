package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Order struct {
	ID                 int64     `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt"`
	FinishCode         int32     `json:"finishCode"`
	IsCompleted        bool      `json:"isCompleted"`
	Destination        Location  `json:"destination"`
	Origin             Location  `json:"origin"`
	DestinationAddress string    `json:"destinationAddress"`
	OriginAddress      string    `json:"originAddress"`
	DeliveryId         int64     `json:"deliveryId"`
	CustomerId         int64     `json:"customerId"`
	DeliveryState      string    `json:"deliveryState"`
	Version            int       `json:"-"`
}

type OrderModel struct {
	DB *sql.DB
}

func (m *OrderModel) UpdateOrder(order *Order) error {
	query := `UPDATE orders Set updated_at = now(),deliveryid= $1, 
                  originlatitude = $2 ,originlongitude= $3,
                  destinationlatitude=  $4,destinationlongitude= $5,
                   delivery = $6 where id = $7`
	args := []any{
		order.DeliveryId,
		order.Origin.Latitude,
		order.Origin.Longitude,
		order.Destination.Latitude,
		order.Destination.Longitude,
		order.DeliveryState,
		order.ID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&order.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}
func (m *OrderModel) GetOrderForUser(user *User) (*Order, error) {
	query := `
	SELECT o.id, created_at, updated_at, finishcode, iscompleted, 
	       originlatitude, originlongitude, 
	       destinationlatitude, destinationlongitude, 
	       deliveryid, destination_adress, origin_adress, delivery, customerid, version
	From orders o 
	    WHERE (deliveryId = $1 OR customerId = $1) AND iscompleted = false AND deliveryid is not null ;`

	args := []any{user.ID}
	var order Order
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&order.ID,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.FinishCode,
		&order.IsCompleted,
		&order.Origin.Latitude,
		&order.Origin.Longitude,
		&order.Destination.Latitude,
		&order.Destination.Longitude,
		&order.DeliveryId,
		&order.DestinationAddress,
		&order.OriginAddress,
		&order.DeliveryState,
		&order.CustomerId,
		&order.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &order, nil
}
func (m *OrderModel) Insert(order *Order) error {
	query := `INSERT INTO orders ( originlatitude, originlongitude, destinationlatitude, 
                    destinationlongitude, 
                    deliveryid, customerid, destination_adress, origin_adress) 
VALUES ($1, $2, $3, $4, $5, $6, $7 , $8) returning id, finishcode`
	args := []any{
		order.Origin.Latitude,
		order.Origin.Longitude,
		order.Destination.Latitude,
		order.Destination.Longitude,
		order.DeliveryId,
		order.CustomerId,
		order.DestinationAddress,
		order.OriginAddress,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var finishCode int
	var orderId int64
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&orderId, &finishCode)
	if err != nil {

		return err
	}
	order.ID = orderId
	order.FinishCode = int32(finishCode)
	return nil
}
func (m *OrderModel) CompleteOrder(id int64) error {
	query := `UPDATE orders Set updated_at = now(), iscompleted = true, 
                  delivery = 'dropped' where id = $1`
	args := []any{
		id,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}
