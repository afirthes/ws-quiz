package main

import (
	"net/http"
)

func (app *application) writeJSONErrorMust(w http.ResponseWriter, status int, message string) {
	if err := writeJSONError(w, status, message); err != nil {
		// If even that fails, log the failure
		app.logger.Errorw("failed to write JSON error response",
			"originalError", err.Error(),
			"writeError", err.Error(),
		)

		// Fallback: plain-text response, best effort
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	app.writeJSONErrorMust(w, http.StatusForbidden, "forbidden")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusConflict, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusNotFound, "not found")
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONErrorMust(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	app.writeJSONErrorMust(w, http.StatusUnauthorized, "unauthorized")
}
