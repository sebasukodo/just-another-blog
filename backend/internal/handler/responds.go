package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type returnError struct {
	Error Error `json:"error"`
}

type Error struct {
	Body string `json:"body"`
	Code int    `json:"code"`
}

type DBError struct {
	Code       int
	Message    string
	LogMessage string
}

func (h *Handler) RespondWithError(w http.ResponseWriter, code int, errorMsg, logMsg string) {

	h.Logger.Error(logMsg)

	respBody := returnError{
		Error: Error{
			Body: errorMsg,
			Code: code,
		},
	}

	h.RespondWithJSON(w, code, respBody)

}

func (h *Handler) RespondWithDatabaseError(w http.ResponseWriter, err error) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			h.RespondWithError(w, 409, "resource already exists", fmt.Sprintf("database resource already exists: %v", err))
			return
		case "23502":
			h.RespondWithError(w, 400, "missing required field", fmt.Sprintf("missing required field for database request: %v", err))
			return
		default:
			h.RespondWithError(w, 500, "internal server error", fmt.Sprintf("database error occured: %v", err))
			return
		}
	}
	h.RespondWithError(w, 500, "internal server error", fmt.Sprintf("database error occured but could not catch specific pqErr: %v", err))
}

func (h *Handler) RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := json.Marshal(payload)
	if err != nil {
		h.Logger.Error(fmt.Sprintf("error marshalling JSON: %s", err))
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	if _, err := w.Write(data); err != nil {
		h.Logger.Error(fmt.Sprintf("error writing response: %v", err))
	}
}

func buildArticleResponse(article database.Article, user database.User, tags []string) RespondArticle {
	if tags == nil {
		tags = []string{}
	}
	return RespondArticle{
		Article: Article{
			Slug:           article.Slug,
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        tags,
			CreatedAt:      article.CreatedAt,
			UpdatedAt:      article.UpdatedAt,
			Favorited:      false,
			FavoritesCount: 0,
			Author: Author{
				Username: user.Username,
				Bio:      nullStringToStringPointer(user.Bio),
				Image:    nullStringToStringPointer(user.Image),
			},
		},
	}
}
