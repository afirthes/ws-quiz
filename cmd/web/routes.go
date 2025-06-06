package main

import (
	"github.com/afirthes/ws-quiz/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func routes(rh *handlers.RestHandlers, wsh *handlers.WsHandlers) http.Handler {
	r := chi.NewRouter()

	r.Get("/", rh.Home)
	r.Get("/ws", wsh.WsEndpoint)

	// Статические файлы
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.Handle("/static/*", fileServer)

	return r
}
