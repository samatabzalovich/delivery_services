package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "8081"

type Config struct {
}

func main() {
	log.Printf("Starting broker sevice on port %s\n", webPort)
	app := Config{}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
