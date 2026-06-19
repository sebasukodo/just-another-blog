package handler

import (
	"database/sql"
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

type buildArticleParams struct {
	Article         database.Article
	User            database.User
	Tags            []string
	IsFavorite      bool
	FavoriteCount   int64
	FollowingAuthor bool
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

	if errors.Is(err, sql.ErrNoRows) {
		h.RespondWithError(w, 404, "resource does not exists", fmt.Sprintf("resource not found: %v", err))
		return
	}

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

func buildArticleResponse(article database.Article, user database.User, tags []string, following bool) RespondArticle {
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
				Username:  user.Username,
				Bio:       nullStringToStringPointer(user.Bio),
				Image:     nullStringToStringPointer(user.Image),
				Following: following,
			},
		},
	}
}

func buildListArticlesResponse(articles []database.ListArticle) RespondArticles {

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
				Bio:       &article.AuthorBio.String,
				Image:     &article.AuthorImage.String,
				Following: article.AuthorIsFollowed,
			},
		})
	}

	return RespondArticles{
		Article:      response,
		ArticleCount: len(articles),
	}
}

func buildArticleFeedResponse(feed []database.FeedArticlesRow) RespondArticles {

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
				Bio:       &article.Bio.String,
				Image:     &article.Image.String,
				Following: true,
			},
		})
	}

	return RespondArticles{
		Article:      response,
		ArticleCount: len(feed),
	}
}

func buildProfileResponse(user database.User, following bool) RespondProfile {
	return RespondProfile{
		Profile: Profile{
			Username:  user.Username,
			Bio:       user.Bio.String,
			Image:     user.Image.String,
			Following: following,
		},
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
			Author: RespondProfile{
				Profile: Profile{
					Username:  comment.Username,
					Bio:       comment.Bio.String,
					Image:     comment.Image.String,
					Following: comment.AuthorIsFollowed,
				},
			},
		})
	}
	return RespondComments{Comments: response}
}
