package main

import (
	"database/sql"
	"delivery_service/internal/data"
	"fmt"
	gosocketio "github.com/graarh/golang-socketio"
	"net/http"
	"sync"

	//socketio "github.com/googollee/go-socket.io"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

const httpPort = "8082"
const grpcPort = "50001"

var counts int64

type Config struct {
	Server      *gosocketio.Server
	Conn        *grpc.ClientConn
	UserDataMap map[string]*data.User
	Models      data.Models
	mutex       sync.RWMutex
}

func main() {

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Panic(err.Error())
	}
	defer conn.Close()
	dbConn := connectToDB()
	if dbConn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	app := Config{
		Conn:        conn,
		Models:      data.New(conn, dbConn),
		UserDataMap: make(map[string]*data.User),
	}
	serv := app.socketServerHandlers()
	app.Server = serv
	//setup http server
	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", app.Server)
	log.Panic(http.ListenAndServe(fmt.Sprintf(":%s", httpPort), serveMux))

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := "postgres://postgres:2529@localhost/postgres"
	//os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}

//srv := &http.Server{
//Addr:    fmt.Sprintf(":%s", webPort),
//Handler: app.routes(),
//}
//
//err = srv.ListenAndServe()
//if err != nil {
//log.Panic(err)
//}
//go func() {
//	if err := app.Server.Serve(); err != nil {
//		log.Fatalf("socketio listen error: %s\n", err)
//	}
//}()
//defer app.Server.Close()
//
//http.Handle("/socket.io/", app.Server)
//
//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil))
