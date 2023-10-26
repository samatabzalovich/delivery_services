package main

import (
	"delivery_service/internal/data"
	"errors"
	"log"
	"math"
	"strconv"
)

var (
	ErrRecordNotFound = errors.New("user not found")
)

func (app *Config) addUserData(socketId string, user *data.User) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.UserDataMap[socketId] = user
}

func (app *Config) removeUserData(socketID string) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	delete(app.UserDataMap, socketID)
}

func (app *Config) sendDeliveryLocationToCustomer(customerSocketId string, message data.Location) error {
	app.mutex.RLock()
	defer app.mutex.RUnlock()

	_, exists := app.UserDataMap[customerSocketId]
	if !exists {
		log.Printf("Recipient user not found")
		return ErrRecordNotFound
	}

	// Use the socket ID of the recipient to send a private message
	app.Server.BroadcastTo(customerSocketId, "message", message)
	return nil
}

func (app *Config) convertIntToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

func (app *Config) findClosestLocation(current data.Location, userDataMap map[string]*data.User) (string, float64) {
	app.mutex.RLock()
	defer app.mutex.RUnlock()
	var closestID string
	var closestDistance float64
	first := true

	for id, loc := range userDataMap {
		distance := app.haversine(current.Latitude, current.Longitude, loc.Location.Latitude, loc.Location.Longitude)
		if first || distance < closestDistance {
			closestID = id
			closestDistance = distance
			first = false
		}
	}

	return closestID, closestDistance
}
func (app *Config) haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1 = lat1 * math.Pi / 180
	lon1 = lon1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	lon2 = lon2 * math.Pi / 180

	// Radius of the Earth in meters
	radius := 6371000.0

	// Haversine formula
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := radius * c

	return distance
}
