package main

import (
	"github.com/afirthes/ws-quiz/internal/env"
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

	app := &application{
		config: config,
	}

	err := app.run(routes())
	if err != nil {
		log.Fatalf("Error starting application: %v", err)
	}

}
