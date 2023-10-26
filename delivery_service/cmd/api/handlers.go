package main

import (
	"delivery_service/internal/data"
	"errors"
	"fmt"
	_ "github.com/googollee/go-socket.io"
	gosocketio "github.com/graarh/golang-socketio"
	transport2 "github.com/graarh/golang-socketio/transport"
	"log"
	"strconv"
)

var (
	errorMethod        = "error"
	noOrderMethod      = "noOrder"
	message            = "message"
	orderExists        = "orderExists"
	deliveryLocation   = "deliveryLocation"
	onLine             = "onLine"
	createOrder        = "createOrder"
	acceptOrder        = "acceptOrder"
	requestToTakeOrder = "requestToTakeOrder"
)

func (app *Config) socketServerHandlers() *gosocketio.Server {
	//create
	server := gosocketio.NewServer(transport2.GetDefaultWebsocketTransport())
	//handle connected
	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Printf("New client connected, id %s", c.Id())
		token := c.RequestHeader().Get("Authorization")
		user, err := app.Models.Users.GetUserForToken(token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				c.Emit(errorMethod, "no user found")

			default:
				c.Emit(errorMethod, "server error")

			}
			return
		}
		user.Channel = c
		order, err := app.Models.Order.GetOrderForUser(user)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				c.Emit(noOrderMethod, "to start to search press the button online")
			default:
				c.Emit(errorMethod, "to start to search press the button online")
			}
		}
		if order != nil {
			c.Emit(orderExists, order)
			c.Join(app.convertIntToString(order.ID))
		}
		c.Emit(message, fmt.Sprintf("Successfully connected to the server %s", strconv.FormatInt(user.ID, 10)))
		//join them to room
		//c.Join("chat")
	})
	server.On(deliveryLocation, func(c *gosocketio.Channel, deliveryMessage data.DeliveryMessage) {
		c.BroadcastTo(app.convertIntToString(deliveryMessage.OrderId), deliveryLocation, deliveryMessage)
	})

	//this method when delivery person is ready to get order, and for changing current delivery location if he is far away from previous one
	server.On(onLine, func(c *gosocketio.Channel, location data.Location) {
		user, exist := app.UserDataMap[c.Id()]
		if exist {
			user.Location = location
			app.mutex.Lock()
			defer app.mutex.Unlock()
			app.UserDataMap[c.Id()] = user
			return
		}
		token := c.RequestHeader().Get("Authorization")
		user, err := app.Models.Users.GetUserForToken(token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				c.Emit(errorMethod, "no user found")

			default:
				c.Emit(errorMethod, "server error")

			}
			return
		}
		user.Location = location
		//add delivery to map
		app.addUserData(c.Id(), user)
	})
	server.On(createOrder, func(c *gosocketio.Channel, order data.Order) error {
		id, _ := app.findClosestLocation(order.Origin, app.UserDataMap)
		err := app.Models.Order.Insert(&order)
		if err != nil {
			c.Emit(errorMethod, err.Error())
			return err
		}
		app.mutex.RLock()
		defer app.mutex.RUnlock()
		app.UserDataMap[id].Channel.Emit(requestToTakeOrder, order)
		c.Join(app.convertIntToString(order.ID))
		return nil
	})
	server.On(acceptOrder, func(c *gosocketio.Channel, order data.Order) error {
		err := app.Models.Order.UpdateOrder(&order)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				c.Emit(errorMethod, "no order found")

			default:
				c.Emit(errorMethod, "server error")
			}
			return err
		}
		c.Join(app.convertIntToString(order.ID))
		app.removeUserData(c.Id())
		return nil
	})
	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Printf("client disconnected id %s", c.Id())
		app.removeUserData(c.Id())
	})

	return server
}
