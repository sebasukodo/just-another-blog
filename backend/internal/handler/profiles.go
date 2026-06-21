package handler

import (
	"fmt"
	"net/http"

	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type RespondProfile struct {
	Profile Profile `json:"profile"`
}

type RespondAuthor struct {
	Profile Profile `json:"author"`
}

type Profile struct {
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {

	username := r.PathValue("username")
	user, err := h.DbQueries.GetUserByUsername(r.Context(), username)
	if err != nil {
		h.RespondWithDatabaseError(w, "profile", err)
		return
	}

	requestUser, authenticated := r.Context().Value(contextKeyUser).(database.User)
	if !authenticated {
		h.RespondWithJSON(w, 200, buildProfileResponse(user, false))
		return
	}

	following, err := h.DbQueries.IsFollowing(r.Context(), database.IsFollowingParams{
		FollowerID:  requestUser.ID,
		FollowingID: user.ID,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, "profile", err)
		return
	}

	h.RespondWithJSON(w, 200, buildProfileResponse(user, following))

}

func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {

	username := r.PathValue("username")
	toFollowUser, err := h.DbQueries.GetUserByUsername(r.Context(), username)
	if err != nil {
		h.RespondWithDatabaseError(w, "profile", err)
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, fieldErrorUser, "missing user context")
		return
	}

	if user.ID == toFollowUser.ID {
		h.RespondWithError(w, 400, fieldErrorUser, fmt.Sprintf("user %v cannot follow himself", user.ID))
		return
	}

	if _, err := h.DbQueries.FollowUser(r.Context(), database.FollowUserParams{
		FollowerID:  user.ID,
		FollowingID: toFollowUser.ID,
	}); err != nil {
		h.RespondWithDatabaseError(w, "profile", err)
		return
	}

	h.RespondWithJSON(w, 200, buildProfileResponse(toFollowUser, true))

}

func (h *Handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {

	username := r.PathValue("username")
	toFollowUser, err := h.DbQueries.GetUserByUsername(r.Context(), username)
	if err != nil {
		h.RespondWithDatabaseError(w, "profile", err)
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, fieldErrorUser, "missing user context")
		return
	}

	if _, err := h.DbQueries.UnfollowUser(r.Context(), database.UnfollowUserParams{
		FollowerID:  user.ID,
		FollowingID: toFollowUser.ID,
	}); err != nil {
		h.RespondWithDatabaseError(w, "profile", err)
		return
	}

	h.RespondWithJSON(w, 200, buildProfileResponse(toFollowUser, false))
}
