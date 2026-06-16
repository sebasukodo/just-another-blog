package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type ArticleCreateRequest struct {
	Article ArticleCreate `json:"article"`
}

type ArticleCreate struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

type RespondArticle struct {
	Article Article `json:"article"`
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
	Username string  `json:"username"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(contextKeyUserID).(uuid.UUID)

	user, err := h.DbQueries.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, "Access Denied", fmt.Sprintf("CreateArticle request failed, no user found for id %v", userID))
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	decoder := json.NewDecoder(r.Body)

	articleInfo := ArticleCreateRequest{}

	if err := decoder.Decode(&articleInfo); err != nil {
		h.RespondWithError(w, 400, "invalid request body", err.Error())
		return
	}

	slug, err := generateSlug(articleInfo.Article.Title)
	if err != nil {
		h.RespondWithError(w, 500, "generating slug failed", err.Error())
		return
	}

	article, err := h.DbQueries.CreateArticle(r.Context(), database.CreateArticleParams{
		AuthorID:    userID,
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

	respBody := RespondArticle{
		Article: Article{
			Slug:           slug,
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        articleInfo.Article.TagList,
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

	h.RespondWithJSON(w, 201, respBody)

}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 404, "Not Found", fmt.Sprintf("GetArticle request failed, no article found for slug %v", slug))
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

	respBody := RespondArticle{
		Article: Article{
			Slug:           slug,
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

	h.RespondWithJSON(w, 200, respBody)

}

func (h *Handler) saveTagsToDatabase(r *http.Request, articleID int64, tags []string) error {
	for _, tag := range tags {
		newTag, err := h.DbQueries.CreateTags(r.Context(), database.CreateTagsParams{
			DisplayName:    tag,
			NormalizedName: strings.ToLower(tag),
		})
		if err != nil {
			return err
		}
		_, err = h.DbQueries.CreateArticleTags(r.Context(), database.CreateArticleTagsParams{
			ArticleID: articleID,
			TagID:     newTag.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
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
