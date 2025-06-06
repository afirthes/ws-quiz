package errors

import (
	"errors"
	"github.com/afirthes/ws-quiz/internal/json"
	"go.uber.org/zap"
	"net/http"
)

var (
	ErrorRenderingTemplate    = errors.New("error rendering template")
	ErrorUserIDRequired       = errors.New("user-name required")
	ErrorUserNameRequired     = errors.New("user-uuid required")
	ErrorUserAlreadyConnected = errors.New("user already connected, more than one connection is not allowed")
)

type ErrorHandler struct {
	log *zap.SugaredLogger
}

func NewErrorHandler(logger *zap.SugaredLogger) *ErrorHandler {
	return &ErrorHandler{
		log: logger,
	}
}

func (eh *ErrorHandler) writeJSONErrorMust(w http.ResponseWriter, status int, message string) {
	if err := json.WriteJSONError(w, status, message); err != nil {
		// If even that fails, log the failure
		eh.log.Errorw("failed to write JSON error response",
			"originalError", err.Error(),
			"writeError", err.Error(),
		)

		// Fallback: plain-text response, best effort
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (eh *ErrorHandler) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	eh.log.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	eh.writeJSONErrorMust(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (eh *ErrorHandler) ForbiddenResponse(w http.ResponseWriter, r *http.Request) {
	eh.log.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	eh.writeJSONErrorMust(w, http.StatusForbidden, "forbidden")
}

func (eh *ErrorHandler) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	eh.log.Warn("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	eh.writeJSONErrorMust(w, http.StatusBadRequest, err.Error())
}

func (eh *ErrorHandler) ConflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	eh.log.Error("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	eh.writeJSONErrorMust(w, http.StatusConflict, err.Error())
}

func (eh *ErrorHandler) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	eh.log.Warn("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	eh.writeJSONErrorMust(w, http.StatusNotFound, "not found")
}

func (eh *ErrorHandler) UnauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	eh.log.Warn("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	eh.writeJSONErrorMust(w, http.StatusUnauthorized, "unauthorized")
}

func (eh *ErrorHandler) UnauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	eh.log.Warn("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	eh.writeJSONErrorMust(w, http.StatusUnauthorized, "unauthorized")
}
