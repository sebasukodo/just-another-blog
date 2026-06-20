package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type AddCommentRequest struct {
	Comment AddComment `json:"comment" validate:"required"`
}

type AddComment struct {
	Body string `json:"body" validate:"required"`
}

type RespondComment struct {
	Comment Comment `json:"comment"`
}

type RespondComments struct {
	Comments []Comment `json:"comments"`
}

type Comment struct {
	Id        int64         `json:"id"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	Body      string        `json:"body"`
	Author    RespondAuthor `json:"author"`
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	decoder := json.NewDecoder(r.Body)

	commentInfo := AddCommentRequest{}

	if err := decoder.Decode(&commentInfo); err != nil {
		h.RespondWithError(w, 400, "body", "invalid request body", err.Error())
		return
	}

	if err := h.Validate.Struct(commentInfo); err != nil {
		h.RespondWithValidationErrors(w, err, "validation failed for creating comment")
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "body", "access denied", "missing user context")
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, "comment", err)
		return
	}

	comment, err := h.DbQueries.AddCommentToArticle(r.Context(), database.AddCommentToArticleParams{
		ArticleID: article.ID,
		AuthorID:  user.ID,
		Body:      commentInfo.Comment.Body,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, "comment", err)
		return
	}

	articleAuthor, err := h.DbQueries.GetUserByID(r.Context(), article.AuthorID)
	if err != nil {
		h.RespondWithDatabaseError(w, "comment", err)
		return
	}

	following, err := h.DbQueries.IsFollowing(r.Context(), database.IsFollowingParams{
		FollowerID:  user.ID,
		FollowingID: articleAuthor.ID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, "comment", err)
		return
	}

	responseBody := RespondComment{
		Comment: Comment{
			Id:        comment.ID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Body:      comment.Body,
			Author:    buildAuthorResponse(user, following),
		},
	}

	h.RespondWithJSON(w, 201, responseBody)

}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, "comment", err)
		return
	}

	userID, isAuthenticated := r.Context().Value(contextKeyUserID).(uuid.UUID)
	if !isAuthenticated {
		userID = uuid.Nil
	}

	allComments, err := h.DbQueries.GetCommentsFromArticle(r.Context(), database.GetCommentsFromArticleParams{
		ArticleID:  article.ID,
		FollowerID: userID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, "comment", err)
		return
	}

	h.RespondWithJSON(w, 200, buildCommentsResponse(allComments))

}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {

	commentIDString := r.PathValue("commentID")
	commentID, err := strconv.ParseInt(commentIDString, 10, 64)
	if err != nil {
		h.RespondWithError(w, 400, "body", "bad request", fmt.Sprintf("could not convert comment ID to int64: %v - %v", commentIDString, err))
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "body", "access denied", "missing user context")
		return
	}

	_, err = h.DbQueries.DeleteCommentFromArticle(r.Context(), database.DeleteCommentFromArticleParams{
		ID:       commentID,
		AuthorID: user.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 403, "body", "forbidden", fmt.Sprintf("could not delete comment %v - not author or not found", commentID))
		} else {
			h.RespondWithDatabaseError(w, "comment", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
