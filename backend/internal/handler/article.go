package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type CreateArticleRequest struct {
	Article CreateArticle `json:"article"`
}

type CreateArticle struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

type UpdateArticleRequest struct {
	Article UpdateArticle `json:"article"`
}

type UpdateArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

type RespondArticle struct {
	Article Article `json:"article"`
}

type RespondArticles struct {
	Article      []Article `json:"articles"`
	ArticleCount int       `json:"articlesCount"`
}

type Article struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int64     `json:"favoritesCount"`
	Author         Author    `json:"author"`
}

type Author struct {
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	decoder := json.NewDecoder(r.Body)

	articleInfo := CreateArticleRequest{}

	if err := decoder.Decode(&articleInfo); err != nil {
		h.RespondWithError(w, 400, "invalid request body", err.Error())
		return
	}

	slug, err := h.generateUniqueSlug(r.Context(), articleInfo.Article.Title, 0)
	if err != nil {
		h.RespondWithError(w, 500, "generating slug failed", err.Error())
		return
	}

	article, err := h.DbQueries.CreateArticle(r.Context(), database.CreateArticleParams{
		AuthorID:    user.ID,
		Slug:        slug,
		Title:       articleInfo.Article.Title,
		Description: articleInfo.Article.Description,
		Body:        articleInfo.Article.Body,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	if err := h.saveTagsToDatabase(r, article.ID, articleInfo.Article.TagList); err != nil {
		h.RespondWithError(w, 201, "article created successfully, but could not store tags", fmt.Sprintf("Article %v created, but could not add tags %v: %v", article.ID, articleInfo.Article.TagList, err))
		return
	}

	h.RespondWithJSON(w, 201, buildArticleResponse(article, user, articleInfo.Article.TagList, false))

}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 404, "not found", fmt.Sprintf("GetArticle request failed, no article found for slug %v", slug))
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

	user, err := h.DbQueries.GetUserByID(r.Context(), article.AuthorID)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	requestUser, authenticated := r.Context().Value(contextKeyUser).(database.User)
	if !authenticated {
		h.RespondWithJSON(w, 200, buildArticleResponse(article, user, tags, false))
		return
	}

	following, err := h.DbQueries.IsFollowing(r.Context(), database.IsFollowingParams{
		FollowerID:  requestUser.ID,
		FollowingID: user.ID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	h.RespondWithJSON(w, 200, buildArticleResponse(article, user, tags, following))

}

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	decoder := json.NewDecoder(r.Body)

	articleInfo := UpdateArticleRequest{}

	if err := decoder.Decode(&articleInfo); err != nil {
		h.RespondWithError(w, 401, "invalid request body", err.Error())
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	if user.ID != article.AuthorID {
		h.RespondWithError(w, 403, "access denied", fmt.Sprintf("UpdateArticle request failed, user %v is not author %v", user.ID.String(), article.AuthorID.String()))
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

	updateInfo := database.UpdateArticleBySlugParams{
		Slug: slug,
	}

	noUpdate := true
	if articleInfo.Article.Title != "" && articleInfo.Article.Title != article.Title {
		updateInfo.Title = stringToNullString(articleInfo.Article.Title)
		newSlug, err := h.generateUniqueSlug(r.Context(), articleInfo.Article.Title, 0)
		if err != nil {
			h.RespondWithError(w, 401, "could not create unique slug for new title", err.Error())
			return
		}
		updateInfo.NewSlug = stringToNullString(newSlug)
		noUpdate = false
	}
	if articleInfo.Article.Body != "" && articleInfo.Article.Body != article.Body {
		updateInfo.Body = stringToNullString(articleInfo.Article.Body)
		noUpdate = false
	}
	if articleInfo.Article.Description != "" && articleInfo.Article.Description != article.Description {
		updateInfo.Description = stringToNullString(articleInfo.Article.Description)
		noUpdate = false
	}

	if noUpdate {
		h.RespondWithJSON(w, 200, buildArticleResponse(article, user, tags, false))
		return
	}

	updatedArticle, err := h.DbQueries.UpdateArticleBySlug(r.Context(), updateInfo)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	h.RespondWithJSON(w, 200, buildArticleResponse(updatedArticle, user, tags, false))

}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	if user.ID != article.AuthorID {
		h.RespondWithError(w, 403, "access denied", fmt.Sprintf("DeleteArticle request failed, user %v is not author %v", user.ID.String(), article.AuthorID.String()))
		return
	}

	if err := h.DbQueries.DeleteArticleById(r.Context(), article.ID); err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (h *Handler) FeedArticles(w http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	limit, offset, err := getFeedArticlesQueries(r)
	if err != nil {
		h.RespondWithError(w, 400, "bad request", "could not parse limit or offset query parameters to int")
		return
	}

	feed, err := h.DbQueries.FeedArticles(r.Context(), database.FeedArticlesParams{
		FollowerID: user.ID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	h.RespondWithJSON(w, 200, buildArticleFeedResponse(feed))

}

func (h *Handler) generateUniqueSlug(ctx context.Context, title string, currentID int64) (string, error) {
	baseSlug, err := generateSlug(title)
	if err != nil {
		return "", err
	}

	slug := baseSlug
	for i := 2; ; i++ {

		existing, err := h.DbQueries.GetArticleBySlug(ctx, slug)
		if err != nil {
			break
		}

		if existing.ID == currentID {
			break
		}

		slug = fmt.Sprintf("%v-%v", baseSlug, i)

	}

	return slug, nil
}

func generateSlug(title string) (string, error) {
	result := strings.ToLower(title)
	result = strings.ReplaceAll(result, " ", "-")
	reg, err := regexp.Compile("[^a-z0-9-]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(result, ""), nil
}

func getFeedArticlesQueries(r *http.Request) (int32, int32, error) {
	var limit int64 = 20
	var offset int64 = 0
	var err error

	limitQuery := r.URL.Query().Get("limit")
	if limitQuery != "" {
		limit, err = strconv.ParseInt(limitQuery, 10, 64)
		if err != nil {
			return int32(limit), int32(offset), err
		}
	}

	offsetQuery := r.URL.Query().Get("offset")
	if offsetQuery != "" {
		offset, err = strconv.ParseInt(offsetQuery, 10, 64)
		if err != nil {
			return int32(limit), int32(offset), err
		}
	}

	return int32(limit), int32(offset), nil
}
