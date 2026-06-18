package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type CommentRequestBody struct {
	Comment CommentRequest `json:"comment"`
}

type CommentRequest struct {
	Body string `json:"body"`
}

type CommentResponseBody struct {
	Comment CommentResponse `json:"comment"`
}

type CommentResponse struct {
	Id        int64               `json:"id"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
	Body      string              `json:"body"`
	Author    ProfileResponseBody `json:"author"`
}

type RespondArticleComments struct {
	Comments []CommentResponse `json:"comments"`
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	decoder := json.NewDecoder(r.Body)

	commentInfo := CommentRequestBody{}

	if err := decoder.Decode(&commentInfo); err != nil {
		h.RespondWithError(w, 400, "invalid request body", err.Error())
		return
	}

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	comment, err := h.DbQueries.AddCommentToArticle(r.Context(), database.AddCommentToArticleParams{
		ArticleID: article.ID,
		AuthorID:  user.ID,
		Body:      commentInfo.Comment.Body,
	})

	articleAuthor, err := h.DbQueries.GetUserByID(r.Context(), article.AuthorID)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	following, err := h.DbQueries.IsFollowing(r.Context(), database.IsFollowingParams{
		FollowerID:  user.ID,
		FollowingID: articleAuthor.ID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	responseBody := CommentResponseBody{
		Comment: CommentResponse{
			Id:        comment.ID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Body:      comment.Body,
			Author:    buildProfileResponse(user, following),
		},
	}

	h.RespondWithJSON(w, 201, responseBody)

}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("slug")

	article, err := h.DbQueries.GetArticleBySlug(r.Context(), slug)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	requestUser, authenticated := r.Context().Value(contextKeyUser).(database.User)
	response, err := h.buildCommentsResponse(r, article, requestUser.ID, authenticated)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}
	h.RespondWithJSON(w, 200, response)

}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {

	commentIDString := r.PathValue("commentID")
	commentID, err := strconv.ParseInt(commentIDString, 10, 64)
	if err != nil {
		h.RespondWithError(w, 400, "bad request", fmt.Sprintf("could not convert comment ID to int64: %v - %v", commentIDString, err))
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	_, err = h.DbQueries.DeleteCommentFromArticle(r.Context(), database.DeleteCommentFromArticleParams{
		ID:       commentID,
		AuthorID: user.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 403, "forbidden", "not author or not found")
		} else {
			h.RespondWithDatabaseError(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
