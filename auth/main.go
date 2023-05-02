package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/eminetto/poc-dapr/auth/security"
	"github.com/eminetto/poc-dapr/auth/user"
	"github.com/eminetto/poc-dapr/auth/user/mysql"
	"github.com/go-chi/httplog"
	"net/http"
	"os"
	"time"

	"context"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
)

const (
	pubsubComponentName = "auditpubsub"
	pubsubTopic         = "audit"
)

func main() {
	// Logger
	logger := httplog.NewLogger("auth", httplog.Options{
		JSON: true,
	})

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		logger.Panic().Msg(err.Error())
	}
	defer db.Close()

	ctx := context.Background()

	repo := mysql.NewUserMySQL(db)
	uService := user.NewService(repo)

	daprClient, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer daprClient.Close()

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Post("/v1/auth", userAuth(ctx, uService, daprClient))
	r.Post("/v1/validate-token", validateToken(ctx, daprClient))

	http.Handle("/", r)
	srv := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         ":" + os.Getenv("PORT"),
		Handler:      http.DefaultServeMux,
	}
	err = srv.ListenAndServe()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}
}

func userAuth(ctx context.Context, uService user.UseCase, daprClient dapr.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oplog := httplog.LogEntry(r.Context())
		var param struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&param)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			oplog.Error().Msg(err.Error())
			return
		}
		err = uService.ValidateUser(ctx, param.Email, param.Password)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			oplog.Error().Msg(err.Error())
			return
		}
		var result struct {
			Token string `json:"token"`
		}
		result.Token, err = security.NewToken(param.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			oplog.Error().Msg(err.Error())
			return
		}

		action := `{"action":"new token generated"}`

		err = daprClient.PublishEvent(ctx, pubsubComponentName, pubsubTopic, []byte(action))
		if err != nil {
			panic(err)
		}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusBadGateway)
			oplog.Error().Msg(err.Error())
			return
		}
		return
	}
}

func validateToken(ctx context.Context, daprClient dapr.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oplog := httplog.LogEntry(r.Context())
		var param struct {
			Token string `json:"token"`
		}
		err := json.NewDecoder(r.Body).Decode(&param)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			oplog.Error().Msg(err.Error())
			return
		}

		t, err := security.ParseToken(param.Token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			oplog.Error().Msg(err.Error())
			return
		}
		tData, err := security.GetClaims(t)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			oplog.Error().Msg(err.Error())
			return
		}
		var result struct {
			Email string `json:"email"`
		}
		result.Email = tData["email"].(string)

		action := `{"action":"token validated:` + result.Email + `"}`

		err = daprClient.PublishEvent(ctx, pubsubComponentName, pubsubTopic, []byte(action))
		if err != nil {
			panic(err)
		}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusBadGateway)
			oplog.Error().Msg(err.Error())
			return
		}
		return
	}
}
