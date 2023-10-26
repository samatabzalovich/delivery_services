package main

import (
	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"log"
	"runtime"
	"time"
)

type Channel struct {
	Channel string `json:"channel"`
}

var (
	errorMethod      = "error"
	noOrderMethod    = "noOrder"
	message          = "message"
	orderExists      = "orderExists"
	deliveryLocation = "deliveryLocation"
	customerLocation = "customerLocation"
)

type Location struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}

func sendJoin(c *gosocketio.Client) {
	log.Println("Acking /join")
	result, err := c.Ack("/join", Channel{"main"}, time.Second*5)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Ack result to /join: ", result)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	c, err := gosocketio.Dial(
		gosocketio.GetUrl("localhost", 8082, false),
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		log.Fatal(err)
	}

	err = c.On("message", func(h *gosocketio.Channel, args string) {
		log.Println("--- Got chat message: ", args)
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		log.Fatal("Disconnected")
	})
	if err != nil {
		log.Fatal(err)
	}
	err = c.On(deliveryLocation, func(h *gosocketio.Channel, args string) {
		log.Println("--- location from server: ", args)
	})
	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("Connected")
		h.Emit("message", "message from client")
		h.BroadcastTo(h.Id(), deliveryLocation, Location{
			Latitude:  0.232323,
			Longitude: 0.2323,
		})
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("5 seconds 4 seconds")
	time.Sleep(60 * time.Second)

	c.Close()

	log.Println(" [x] Complete")
}
