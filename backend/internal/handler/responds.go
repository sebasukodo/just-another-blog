package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type returnError struct {
	Error string `json:"error"`
}

func (h *Handler) RespondWithError(w http.ResponseWriter, code int, errorMsg, logMsg string) {

	h.Logger.Error(logMsg)

	respBody := returnError{
		Error: errorMsg,
	}

	h.RespondWithJSON(w, code, respBody)

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
