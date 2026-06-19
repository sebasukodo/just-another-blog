package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

func (h *Handler) FavoriteArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 404, "not found", fmt.Sprintf("FavoriteArticle request failed, no article found for slug %v", slug))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	tags, err := h.DbQueries.GetArticleTagsByArticleID(r.Context(), article.ID)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	if tags == nil {
		tags = []string{}
	}

	_, err = h.DbQueries.FavoriteArticle(r.Context(), database.FavoriteArticleParams{
		ArticleID: article.ID,
		UserID:    user.ID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	article.FavoriteCount += 1

	following, err := h.DbQueries.IsFollowing(r.Context(), database.IsFollowingParams{
		FollowerID:  user.ID,
		FollowingID: article.AuthorID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	h.RespondWithJSON(w, 201, buildArticleResponse(article, tags, following, true))

}

func (h *Handler) UnfavoriteArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 404, "not found", fmt.Sprintf("UnfavoriteArticle request failed, no article found for slug %v", slug))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	tags, err := h.DbQueries.GetArticleTagsByArticleID(r.Context(), article.ID)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	if tags == nil {
		tags = []string{}
	}

	if err = h.DbQueries.UnfavoriteArticle(r.Context(), database.UnfavoriteArticleParams{
		ArticleID: article.ID,
		UserID:    user.ID,
	}); err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	article.FavoriteCount -= 1

	following, err := h.DbQueries.IsFollowing(r.Context(), database.IsFollowingParams{
		FollowerID:  user.ID,
		FollowingID: article.AuthorID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	h.RespondWithJSON(w, 200, buildArticleResponse(article, tags, following, false))

}
