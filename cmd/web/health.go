package main

import (
	"github.com/afirthes/ws-quiz/internal/json"
	"net/http"
)

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"Env":     app.Config.Env,
		"version": version,
	}

	if err := json.JsonResponse(w, http.StatusOK, data); err != nil {
		app.ErrorHandler.InternalServerError(w, r, err)
	}
}
