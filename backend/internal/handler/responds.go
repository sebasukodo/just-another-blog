package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

const (
	fieldErrorArticle string = "article"
	fieldErrorComment string = "comment"
	fieldErrorProfile string = "profile"
	fieldErrorUser    string = "user"
	fieldErrorToken   string = "token"
)

func (h *Handler) RespondWithError(w http.ResponseWriter, code int, field, logMsg string) {
	h.Logger.Error(logMsg)

	errMsg := ""
	switch code {
	case 400:
		errMsg = "can't be empty"
	case 401:
		if field == fieldErrorToken {
			errMsg = "is missing"
		} else {
			errMsg = "not authorized"
		}
	case 403:
		errMsg = "forbidden"
	case 404:
		errMsg = "not found"
	case 409:
		if field == fieldErrorUser {
			field = "username"
			errMsg = "has already been taken"
		} else {
			errMsg = "already exists"
		}
	case 422:
		errMsg = "unprocessable entity"
	case 500:
		errMsg = "internal server error"
	default:
		errMsg = "some error occured"
	}

	respBody := GenericErrorModel{
		Errors: map[string][]string{
			field: {errMsg},
		},
	}

	h.RespondWithJSON(w, code, respBody)
}

func (h *Handler) RespondWithValidationErrors(w http.ResponseWriter, err error, logMsg string) {
	h.Logger.Error(logMsg)

	errorMap := make(map[string][]string)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			field := lowerFirst(fieldError.Field())
			msg := validationMessage(fieldError)
			errorMap[field] = append(errorMap[field], msg)
		}
	} else {
		errorMap["body"] = []string{err.Error()}
	}

	respBody := GenericErrorModel{Errors: errorMap}
	h.RespondWithJSON(w, http.StatusUnprocessableEntity, respBody)
}

func (h *Handler) RespondWithDatabaseError(w http.ResponseWriter, field string, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		h.RespondWithError(w, 404, field, fmt.Sprintf("resource not found: %v", err))
		return
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			h.RespondWithError(w, 409, field, fmt.Sprintf("database resource already exists: %v", err))
			return
		case "23502":
			h.RespondWithError(w, 400, field, fmt.Sprintf("missing required field for database request: %v", err))
			return
		default:
			h.RespondWithError(w, 500, field, fmt.Sprintf("database error occured: %v", err))
			return
		}
	}

	h.RespondWithError(w, 500, field, fmt.Sprintf("database error occured but could not catch specific pqErr: %v", err))
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

func buildArticleResponse(article database.GetArticleBySlugRow, following, isFavorite bool) RespondArticle {
	return RespondArticle{
		Article: Article{
			Slug:           article.Slug,
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        article.Tags,
			CreatedAt:      article.CreatedAt,
			UpdatedAt:      article.UpdatedAt,
			Favorited:      isFavorite,
			FavoritesCount: article.FavoriteCount,
			Author: Author{
				Username:  article.Username,
				Bio:       nullStringToStringPointer(article.Bio),
				Image:     nullStringToStringPointer(article.Image),
				Following: following,
			},
		},
	}
}

func buildListArticlesResponse(articles []database.ListArticle, articleCount int64) RespondArticles {

	response := []ArticleNoBody{}

	for _, article := range articles {

		response = append(response, ArticleNoBody{
			Slug:           article.Slug,
			Title:          article.Title,
			Description:    article.Description,
			TagList:        article.Tags,
			CreatedAt:      article.CreatedAt,
			UpdatedAt:      article.UpdatedAt,
			Favorited:      article.IsFavorited,
			FavoritesCount: article.FavoritesCount,
			Author: Author{
				Username:  article.AuthorUsername,
				Bio:       nullStringToStringPointer(article.AuthorBio),
				Image:     nullStringToStringPointer(article.AuthorImage),
				Following: article.AuthorIsFollowed,
			},
		})
	}

	return RespondArticles{
		Article:      response,
		ArticleCount: articleCount,
	}
}

func buildArticleFeedResponse(feed []database.FeedArticlesRow, feedCount int64) RespondArticles {

	response := []ArticleNoBody{}

	for _, article := range feed {

		response = append(response, ArticleNoBody{
			Slug:           article.Slug,
			Title:          article.Title,
			Description:    article.Description,
			TagList:        article.Tags,
			CreatedAt:      article.CreatedAt,
			UpdatedAt:      article.UpdatedAt,
			Favorited:      article.IsFavorited,
			FavoritesCount: article.FavoritesCount,
			Author: Author{
				Username:  article.Username,
				Bio:       nullStringToStringPointer(article.Bio),
				Image:     nullStringToStringPointer(article.Image),
				Following: true,
			},
		})
	}

	return RespondArticles{
		Article:      response,
		ArticleCount: feedCount,
	}
}

func buildProfileResponse(user database.User, following bool) RespondProfile {
	return RespondProfile{
		Profile: Profile{
			Username:  user.Username,
			Bio:       nullStringToStringPointer(user.Bio),
			Image:     nullStringToStringPointer(user.Image),
			Following: following,
		},
	}
}

func buildAuthorResponse(user database.User, following bool) Author {
	return Author{
		Username:  user.Username,
		Bio:       nullStringToStringPointer(user.Bio),
		Image:     nullStringToStringPointer(user.Image),
		Following: following,
	}
}

func buildCommentsResponse(allComments []database.GetCommentsFromArticleRow) RespondComments {
	response := []Comment{}
	for _, comment := range allComments {
		response = append(response, Comment{
			Id:        comment.ID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Body:      comment.Body,
			Author: Author{
				Username:  comment.Username,
				Bio:       nullStringToStringPointer(comment.Bio),
				Image:     nullStringToStringPointer(comment.Image),
				Following: comment.AuthorIsFollowed,
			},
		})
	}
	return RespondComments{Comments: response}
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
