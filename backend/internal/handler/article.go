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

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

const queryLimit int64 = 150

type CreateArticleRequest struct {
	Article CreateArticle `json:"article" validate:"required"`
}

type CreateArticle struct {
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Body        string   `json:"body" validate:"required"`
	TagList     []string `json:"tagList"`
}

type UpdateArticleRequest struct {
	Article UpdateArticle `json:"article" validate:"required"`
}

type UpdateArticle struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

type RespondArticle struct {
	Article Article `json:"article"`
}

type RespondArticles struct {
	Article      []ArticleNoBody `json:"articles"`
	ArticleCount int64           `json:"articlesCount"`
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

type ArticleNoBody struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
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

	decoder := json.NewDecoder(r.Body)

	articleInfo := CreateArticleRequest{}

	if err := decoder.Decode(&articleInfo); err != nil {
		h.RespondWithError(w, 422, fieldErrorArticle, err.Error())
		return
	}

	if err := h.Validate.Struct(articleInfo); err != nil {
		h.RespondWithValidationErrors(w, err, "validation failed for creating article")
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, fieldErrorArticle, "missing user context")
		return
	}

	slug, err := h.generateUniqueSlug(r.Context(), articleInfo.Article.Title, 0)
	if err != nil {
		h.RespondWithError(w, 500, fieldErrorArticle, err.Error())
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
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	if err := h.saveTagsToDatabase(r, article.ID, articleInfo.Article.TagList); err != nil {
		h.RespondWithError(w, 201, fieldErrorArticle, fmt.Sprintf("Article %v created, but could not add tags %v: %v", article.ID, articleInfo.Article.TagList, err))
		return
	}

	fullArticle, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	h.RespondWithJSON(w, 201, buildArticleResponse(fullArticle, false, false))

}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 404, fieldErrorArticle, fmt.Sprintf("GetArticle request failed, no article found for slug %v", slug))
		} else {
			h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		}
		return
	}

	requestUser, authenticated := r.Context().Value(contextKeyUser).(database.User)
	if !authenticated {
		h.RespondWithJSON(w, 200, buildArticleResponse(article, false, false))
		return
	}

	isFavorite, err := h.DbQueries.GetArticleIsFavorite(r.Context(), database.GetArticleIsFavoriteParams{
		ArticleID: article.ID,
		UserID:    requestUser.ID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	following, err := h.DbQueries.IsFollowing(r.Context(), database.IsFollowingParams{
		FollowerID:  requestUser.ID,
		FollowingID: article.AuthorID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	h.RespondWithJSON(w, 200, buildArticleResponse(article, following, isFavorite))

}

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	decoder := json.NewDecoder(r.Body)

	articleInfo := UpdateArticleRequest{}

	if err := decoder.Decode(&articleInfo); err != nil {
		h.RespondWithError(w, 401, fieldErrorArticle, err.Error())
		return
	}

	if err := h.Validate.Struct(articleInfo); err != nil {
		h.RespondWithValidationErrors(w, err, "validation failed for updating article")
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, fieldErrorArticle, "missing user context")
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	if user.ID != article.AuthorID {
		h.RespondWithError(w, 403, fieldErrorArticle, fmt.Sprintf("UpdateArticle request failed, user %v is not author %v", user.ID.String(), article.AuthorID.String()))
		return
	}

	updateInfo := database.UpdateArticleBySlugParams{
		Slug: slug,
	}

	noUpdate := true
	if articleInfo.Article.Title != "" && articleInfo.Article.Title != article.Title {
		updateInfo.Title = stringToNullString(articleInfo.Article.Title)
		newSlug, err := h.generateUniqueSlug(r.Context(), articleInfo.Article.Title, 0)
		if err != nil {
			h.RespondWithError(w, 401, fieldErrorArticle, err.Error())
			return
		}
		updateInfo.NewSlug = stringToNullString(newSlug)
		noUpdate = false
		slug = newSlug
	}
	if articleInfo.Article.Body != "" && articleInfo.Article.Body != article.Body {
		updateInfo.Body = stringToNullString(articleInfo.Article.Body)
		noUpdate = false
	}
	if articleInfo.Article.Description != "" && articleInfo.Article.Description != article.Description {
		updateInfo.Description = stringToNullString(articleInfo.Article.Description)
		noUpdate = false
	}
	updateTags := articleInfo.Article.TagList != nil
	if updateTags {
		noUpdate = false
	}

	if noUpdate {
		h.RespondWithJSON(w, 200, buildArticleResponse(article, false, false))
		return
	}

	// starting transaction
	tx, err := h.Db.BeginTx(r.Context(), nil)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}
	defer tx.Rollback()
	qtx := h.DbQueries.WithTx(tx)

	contentChanged := updateInfo.Title.Valid || updateInfo.Body.Valid || updateInfo.Description.Valid
	if contentChanged {
		_, err = qtx.UpdateArticleBySlug(r.Context(), updateInfo)
		if err != nil {
			h.RespondWithDatabaseError(w, fieldErrorArticle, err)
			return
		}
	}

	if updateTags {
		if err := qtx.DeleteArticleTagsByArticleID(r.Context(), article.ID); err != nil {
			h.RespondWithDatabaseError(w, fieldErrorArticle, err)
			return
		}

		noDuplication := make(map[string]struct{})
		for _, tagName := range articleInfo.Article.TagList {
			if _, ok := noDuplication[tagName]; ok {
				continue
			}
			noDuplication[tagName] = struct{}{}

			tag, err := qtx.CreateTags(r.Context(), tagName)
			if err != nil {
				h.RespondWithDatabaseError(w, fieldErrorArticle, err)
				return
			}
			_, err = qtx.CreateArticleTags(r.Context(), database.CreateArticleTagsParams{
				ArticleID: article.ID,
				TagID:     tag.ID,
			})
			if err != nil {
				h.RespondWithDatabaseError(w, fieldErrorArticle, err)
				return
			}
		}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	fullArticle, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	h.RespondWithJSON(w, 200, buildArticleResponse(fullArticle, false, false))

}

func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, fieldErrorArticle, "missing user context")
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	if user.ID != article.AuthorID {
		h.RespondWithError(w, 403, fieldErrorArticle, fmt.Sprintf("DeleteArticle request failed, user %v is not author %v", user.ID.String(), article.AuthorID.String()))
		return
	}

	if err := h.DbQueries.DeleteArticleById(r.Context(), article.ID); err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	w.WriteHeader(204)

}

func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {

	limit, offset, err := getListArticlesQueries(r)
	if err != nil {
		h.RespondWithError(w, 400, fieldErrorArticle, fmt.Sprintf("could not parse limit or offset query parameters to int: %v", err))
		return
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	queryBuilder := psql.Select(
		"a.id", "a.created_at", "a.updated_at", "a.author_id",
		"a.slug", "a.title", "a.description",
		"author.username", "author.bio", "author.image",
		"array_agg(t.name ORDER BY t.name) FILTER (WHERE t.name IS NOT NULL)::text[] AS tags",
	)

	userID, isAuthenticated := r.Context().Value(contextKeyUserID).(uuid.UUID)
	if isAuthenticated {
		queryBuilder = queryBuilder.Column(sq.Expr(
			"EXISTS (SELECT 1 FROM user_follows uf WHERE uf.follower_id = ? AND uf.following_id = a.author_id) AS author_is_followed",
			userID.String(),
		))
		queryBuilder = queryBuilder.Column(sq.Expr(
			"EXISTS (SELECT 1 FROM article_favorites af2 WHERE af2.article_id = a.id AND af2.user_id = ?) AS is_favorited",
			userID.String(),
		))
	} else {
		queryBuilder = queryBuilder.Column("false AS author_is_followed")
		queryBuilder = queryBuilder.Column("false AS is_favorited")
	}

	queryBuilder = queryBuilder.Column("(SELECT COUNT(*) FROM article_favorites af WHERE af.article_id = a.id) AS favorites_count")

	queryBuilder = queryBuilder.
		From("articles a").
		LeftJoin("users author ON author.id = a.author_id").
		LeftJoin("article_tags at ON at.article_id = a.id").
		LeftJoin("tags t ON t.id = at.tag_id")

	queryBuilder = applyArticleFiltersToQuery(queryBuilder, r)

	queryBuilder = queryBuilder.
		GroupBy("a.id", "author.username", "author.bio", "author.image").
		OrderBy("a.created_at DESC").
		Limit(limit).
		Offset(offset)

	sqlString, args, err := queryBuilder.ToSql()
	if err != nil {
		h.RespondWithError(w, 400, fieldErrorArticle, fmt.Sprintf("an error occured while listing articles: %v", err))
		return
	}

	// #nosec G701 -- sqlString is built via Squirrel's query builder
	// user input is passed as args and not interpolated into SQL string
	rows, err := h.Db.QueryContext(r.Context(), sqlString, args...)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	articles, err := database.ScanListArticles(rows)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	articlesCount, err := h.CountListArticles(w, r)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	h.RespondWithJSON(w, 200, buildListArticlesResponse(articles, articlesCount))
}

func (h *Handler) CountListArticles(w http.ResponseWriter, r *http.Request) (int64, error) {

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	queryBuilder := psql.Select(
		"COUNT(DISTINCT a.id)",
	)

	queryBuilder = queryBuilder.
		From("articles a").
		LeftJoin("users author ON author.id = a.author_id").
		LeftJoin("article_tags at ON at.article_id = a.id").
		LeftJoin("tags t ON t.id = at.tag_id")

	queryBuilder = applyArticleFiltersToQuery(queryBuilder, r)

	sqlString, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, err
	}

	// #nosec G701 -- sqlString is built via Squirrel's query builder
	// user input is passed as args and not interpolated into SQL string
	row := h.Db.QueryRowContext(r.Context(), sqlString, args...)
	var count int64
	err = row.Scan(&count)
	return count, err
}

func (h *Handler) FeedArticles(w http.ResponseWriter, r *http.Request) {

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, fieldErrorArticle, "missing user context")
		return
	}

	limit, offset, err := getFeedArticlesQueries(r)
	if err != nil {
		h.RespondWithError(w, 400, fieldErrorArticle, fmt.Sprintf("could not parse limit or offset query parameters to int: %v", err))
		return
	}

	followCount, err := h.DbQueries.GetUserFollowCount(r.Context(), user.ID)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}
	if followCount == 0 {
		h.RespondWithJSON(w, 200, buildArticleFeedResponse(nil, 0))
		return
	}

	feed, err := h.DbQueries.FeedArticles(r.Context(), database.FeedArticlesParams{
		UserID: user.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	feedCount, err := h.DbQueries.GetArticlesFeedCount(r.Context(), user.ID)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorArticle, err)
		return
	}

	h.RespondWithJSON(w, 200, buildArticleFeedResponse(feed, feedCount))

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

func parseLimitOffset(r *http.Request) (int64, int64, error) {
	var limit int64 = 20
	var offset int64 = 0
	var err error

	limitQuery := r.URL.Query().Get("limit")
	if limitQuery != "" {
		limit, err = strconv.ParseInt(limitQuery, 10, 32)
		if err != nil || limit < 0 {
			return 0, 0, fmt.Errorf("limit %v - error %v", limitQuery, err)
		}
		if limit > queryLimit {
			limit = queryLimit
		}
	}

	offsetQuery := r.URL.Query().Get("offset")
	if offsetQuery != "" {
		offset, err = strconv.ParseInt(offsetQuery, 10, 32)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("offset %v - error %v", offsetQuery, err)
		}
	}
	return limit, offset, nil
}

func getFeedArticlesQueries(r *http.Request) (int32, int32, error) {
	limit, offset, err := parseLimitOffset(r)
	// #nosec G115 -- strconv.ParseInt(..., 10, 32) limits the value to int32 range
	return int32(limit), int32(offset), err
}

func getListArticlesQueries(r *http.Request) (uint64, uint64, error) {
	limit, offset, err := parseLimitOffset(r)
	// #nosec G115 -- strconv.ParseInt(..., 10, 32) limits the value to int32 range and
	// non-negative validation guarantees safe conversion to uint64
	return uint64(limit), uint64(offset), err
}

func applyArticleFiltersToQuery(queryBuilder sq.SelectBuilder, r *http.Request) sq.SelectBuilder {

	tagQuery := stringToNullString(r.URL.Query().Get("tag"))
	if tagQuery.Valid {
		queryBuilder = queryBuilder.Where(sq.Expr(
			`EXISTS (
				SELECT 1 FROM article_tags at2
				JOIN tags t2 ON t2.id = at2.tag_id
				WHERE at2.article_id = a.id AND t2.name = ?
			)`,
			tagQuery.String,
		))
	}

	authorQuery := stringToNullString(r.URL.Query().Get("author"))
	if authorQuery.Valid {
		queryBuilder = queryBuilder.Where(sq.Eq{"author.username": authorQuery.String})
	}

	favQuery := stringToNullString(r.URL.Query().Get("favorited"))
	if favQuery.Valid {
		queryBuilder = queryBuilder.Where(sq.Expr(
			`EXISTS (
            SELECT 1 FROM article_favorites af3
            JOIN users u3 ON u3.id = af3.user_id
            WHERE af3.article_id = a.id AND u3.username = ?
        )`,
			favQuery.String,
		))
	}

	return queryBuilder

}
