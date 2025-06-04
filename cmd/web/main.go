package main

import (
	"github.com/afirthes/ws-quiz/internal/env"
	"github.com/afirthes/ws-quiz/internal/errors"
	"github.com/afirthes/ws-quiz/internal/handlers"
	"go.uber.org/zap"
	"log"
)

var version = "0.0.1"

type config struct {
	addr string
	env  string
	rdb  rdbConfig
}

type rdbConfig struct {
	addr string
}

func main() {

	config := config{
		addr: env.GetString("APP_ADDR", ":8080"),
		rdb: rdbConfig{
			addr: env.GetString("REDIS_ADDR", "localhost:6379"),
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Error handler
	errorHandler := errors.NewErrorHandler(logger)

	app := &application{
		config:       config,
		logger:       logger,
		errorHandler: errorHandler,
	}

	handlers := handlers.NewRestHandlers(logger, errorHandler)

	err := app.run(routes(handlers))
	if err != nil {
		log.Fatalf("Error starting application: %v", err)
	}

}
