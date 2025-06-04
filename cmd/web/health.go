package main

import (
	"github.com/afirthes/ws-quiz/internal/json"
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	if err := json.JsonResponse(w, http.StatusOK, data); err != nil {
		app.errorHandler.InternalServerError(w, r, err)
	}
}
