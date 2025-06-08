package main

import (
	"database/sql"
	"github.com/afirthes/ws-quiz/internal/env"
	"github.com/afirthes/ws-quiz/internal/errors"
	"github.com/afirthes/ws-quiz/internal/handlers"
	"go.uber.org/zap"
	"log"
)

var version = "0.0.1"

type Config struct {
	Addr string
	Env  string
	Rdb  rdbConfig
}

type rdbConfig struct {
	addr string
}

var db *sql.DB

func main() {

	var err error
	db, err = sql.Open("postgres", "postgres://admin:adminpassword@localhost:5432/avitodb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	config := Config{
		Addr: env.GetString("APP_ADDR", ":8080"),
		Rdb: rdbConfig{
			addr: env.GetString("REDIS_ADDR", "localhost:6379"),
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Error handler
	errorHandler := errors.NewErrorHandler(logger)

	app := &Application{
		Config:       config,
		Logger:       logger,
		ErrorHandler: errorHandler,
	}

	restHandlers := handlers.NewRestHandlers(logger, errorHandler)

	err = app.run(routes(restHandlers))
	if err != nil {
		log.Fatalf("Error starting Application: %v", err)
	}

}
