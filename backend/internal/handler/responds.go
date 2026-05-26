package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

type returnError struct {
	Error string `json:"error"`
}

type DBError struct {
	Code       int
	Message    string
	LogMessage string
}

func (h *Handler) RespondWithError(w http.ResponseWriter, code int, errorMsg, logMsg string) {

	h.Logger.Error(logMsg)

	respBody := returnError{
		Error: errorMsg,
	}

	h.RespondWithJSON(w, code, respBody)

}

func (h *Handler) RespondWithDatabaseError(w http.ResponseWriter, err error) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			h.RespondWithError(w, 409, "Resource already exists", fmt.Sprintf("Database resource already exists: %v", err))
			return
		case "23502":
			h.RespondWithError(w, 400, "Missing required field", fmt.Sprintf("Missing required field for database request: %v", err))
			return
		default:
			h.RespondWithError(w, 500, "Internal Server Error", fmt.Sprintf("Database Error occured: %v", err))
			return
		}
	}
	h.RespondWithError(w, 500, "Internal Server Error", fmt.Sprintf("Database Error occured but could not catch specific pqErr: %v", err))
}

func (h *Handler) RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := json.Marshal(payload)
	if err != nil {
		h.Logger.Error(fmt.Sprintf("Error marshalling JSON: %s", err))
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}
