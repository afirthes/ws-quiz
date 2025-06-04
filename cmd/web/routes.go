package main

import (
	"github.com/afirthes/ws-quiz/internal/handlers"
	"github.com/bmizerany/pat"
	"net/http"
)

func routes(rh *handlers.RestHandlers) http.Handler {
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(rh.Home))

	// TODO: enable ws
	//mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))
	return mux
}
