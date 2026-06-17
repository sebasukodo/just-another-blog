package handler

import (
	"fmt"
	"net/http"

	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type ProfileResponseBody struct {
	Profile Profile `json:"profile"`
}

type Profile struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {

	username := r.PathValue("username")
	toFollowUser, err := h.DbQueries.GetUserByUsername(r.Context(), username)
	if err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	user, ok := r.Context().Value(contextKeyUser).(database.User)
	if !ok {
		h.RespondWithError(w, 401, "access denied", "missing user context")
		return
	}

	if user.ID == toFollowUser.ID {
		h.RespondWithError(w, 400, "cannot follow yourself", fmt.Sprintf("user %v cannot follow himself", user.ID))
		return
	}

	if _, err := h.DbQueries.FollowUser(r.Context(), database.FollowUserParams{
		FollowerID:  user.ID,
		FollowingID: toFollowUser.ID,
	}); err != nil {
		h.RespondWithDatabaseError(w, err)
		return
	}

	respBody := ProfileResponseBody{
		Profile: Profile{
			Username:  toFollowUser.Username,
			Bio:       toFollowUser.Bio.String,
			Image:     toFollowUser.Image.String,
			Following: true,
		},
	}

	h.RespondWithJSON(w, 201, respBody)

}
