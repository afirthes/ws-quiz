package main

import (
	"github.com/afirthes/ws-quiz/internal/env"
	"github.com/afirthes/ws-quiz/internal/errors"
	"github.com/afirthes/ws-quiz/internal/handlers"
	"github.com/afirthes/ws-quiz/internal/services"
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

func main() {

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
		UserService:  services.NewUserService(logger),
		QuizService:  services.NewQuizService(logger),
	}

	restHandlers := handlers.NewRestHandlers(logger, errorHandler)
	wsHandlers := handlers.NewWebsocketsHandlers(logger, errorHandler, app.UserService, app.QuizService)

	err := app.run(routes(restHandlers, wsHandlers), wsHandlers)
	if err != nil {
		log.Fatalf("Error starting Application: %v", err)
	}

}
