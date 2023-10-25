package main

import (
	"broker-service/auth"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"time"
)

// RequestPayload describes the JSON that this service accepts as an HTTP Post request

type RequestPayload struct {
	Action string       `json:"action"`
	Auth   AuthPayload  `json:"auth,omitempty"`
	Log    LogPayload   `json:"log,omitempty"`
	Mail   MailPayload  `json:"mail,omitempty"`
	Reg    RegPayload   `json:"reg,omitempty"`
	Token  TokenPayload `json:"token"`
}
type TokenPayload struct {
	Bearer string `json:"bearer"`
}

// MailPayload is the embedded type (in RequestPayload) that describes an email message to be sent
type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// AuthPayload is the embedded type (in RequestPayload) that describes an authentication request
type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegPayload struct {
	UserName string `json:"userName"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

// LogPayload is the embedded type (in RequestPayload) that describes a request to log something
type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.AuthViaGRpc(w, requestPayload)
	case "token":
		app.CreateTokenViaGRpc(w, requestPayload)
	case "reg":
		app.RegViaGRpc(w, requestPayload)
	//case "mail":
	//	app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) CreateTokenViaGRpc(w http.ResponseWriter, r RequestPayload) {

	conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := auth.NewTokenServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.CreateAuthenticationToken(ctx, &auth.TokenRequest{
		AuthEntry: &auth.Auth{
			Email:    r.Auth.Email,
			Password: r.Auth.Password,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = res.Result

	app.writeJSON(w, http.StatusAccepted,
		map[string]any{
			"result": payload,
			"user":   res.User,
		})
}

func (app *Config) RegViaGRpc(w http.ResponseWriter, r RequestPayload) {

	conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := auth.NewRegServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.Register(ctx, &auth.RegRequest{
		RegEntry: &auth.Reg{
			Email:     r.Reg.Email,
			Password:  r.Reg.Password,
			UserName:  r.Reg.UserName,
			Type:      r.Reg.Type,
			Activated: false,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = res.Result

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) AuthViaGRpc(w http.ResponseWriter, r RequestPayload) {

	conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := auth.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.Authenticate(ctx, &auth.AuthRequest{
		TokenEntry: &auth.Token{
			Token: r.Token.Bearer,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("%+v, result: %s", res.User, res.Result)

	app.writeJSON(w, http.StatusAccepted, map[string]any{"exist": res.Result, "user": res.User, "error": payload.Error})
}
