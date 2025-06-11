package main

import (
	"database/sql"
	"github.com/afirthes/ws-quiz/internal/env"
	"github.com/afirthes/ws-quiz/internal/errors"
	"github.com/afirthes/ws-quiz/internal/handlers"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
	"os"
	"path/filepath"
)

var version = "0.0.1"

type Config struct {
	Addr string
	Env  string
}

var db *sql.DB

func main() {

	exe, _ := os.Executable()
	envPath := filepath.Join(filepath.Dir(exe), ".env")
	log.Println("Env file: ", envPath)
	// Загружаем переменные окружения из .env
	if err := godotenv.Load(envPath); err != nil {
		log.Println("⚠️  .env not found, continuing with system env")
	}
	log.Println("Statid dir", env.GetString("STATIC_DIR", "./static/"))

	var err error
	db, err = sql.Open("postgres", "postgres://admin:adminpassword@localhost:5432/avitodb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	config := Config{
		Addr: env.GetString("APP_ADDR", ":8080"),
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
