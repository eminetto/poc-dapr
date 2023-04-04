package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/eminetto/api-o11y/pkg/middleware"
	"github.com/eminetto/api-o11y/votes/vote"
	"github.com/eminetto/api-o11y/votes/vote/mysql"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"time"
)

// @todo get this variables from env or config
const (
	DB_USER     = "votes_user"
	DB_PASSWORD = "votes_pwd"
	DB_HOST     = "localhost"
	DB_DATABASE = "votes_db"
	DB_PORT     = "3308"
	PORT        = "8083"
)

func main() {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_DATABASE)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Panic(err.Error())
	}
	defer db.Close()
	repo := mysql.NewVoteMySQL(db)

	vService := vote.NewService(repo)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(middleware.IsAuthenticated)
	r.Post("/v1/vote", storeVote(vService))

	http.Handle("/", r)
	logger := log.New(os.Stderr, "logger: ", log.Lshortfile)
	srv := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         ":" + PORT,
		Handler:      http.DefaultServeMux,
		ErrorLog:     logger,
	}
	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func storeVote(vService vote.UseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var v vote.Vote
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		v.Email = r.Context().Value("email").(string)
		var result struct {
			ID uuid.UUID `json:"id"`
		}
		result.ID, err = vService.Store(r.Context(), &v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		return
	}
}
