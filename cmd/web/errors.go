package main

import (
	"net/http"
)

func (app *Application) writeJSONErrorMust(w http.ResponseWriter, status int, message string) {
	if err := writeJSONError(w, status, message); err != nil {
		// If even that fails, log the failure
		app.Logger.Errorw("failed to write JSON error response",
			"originalError", err.Error(),
			"writeError", err.Error(),
		)

		// Fallback: plain-text response, best effort
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *Application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *Application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.Logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	app.writeJSONErrorMust(w, http.StatusForbidden, "forbidden")
}

func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warn("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusBadRequest, err.Error())
}

func (app *Application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Error("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusConflict, err.Error())
}

func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warn("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusNotFound, "not found")
}

func (app *Application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warn("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warn("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	app.writeJSONErrorMust(w, http.StatusUnauthorized, "unauthorized")
}
