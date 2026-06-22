package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sebasukodo/just-another-blog/backend/internal/auth"
	"github.com/sebasukodo/just-another-blog/backend/internal/database"
)

type RegisterUserRequest struct {
	User RegisterUser `json:"user" validate:"required"`
}

type RegisterUser struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginUserRequest struct {
	User LoginUser `json:"user" validate:"required"`
}

type LoginUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	User UpdateUser `json:"user" validate:"required"`
}

type UpdateUser struct {
	Username string           `json:"username"`
	Email    string           `json:"email"`
	Password string           `json:"password"`
	Bio      Nullable[string] `json:"bio"`
	Image    Nullable[string] `json:"image"`
}

type RespondUser struct {
	User User `json:"user"`
}

type User struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Token    string  `json:"token"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}

type Nullable[T any] struct {
	Set   bool
	Value *T
}

const JWTExpiresIn = time.Duration(15) * time.Minute
const minPasswordLength = 8

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {

	limitedRequest := http.MaxBytesReader(w, r.Body, 2048)
	decoder := json.NewDecoder(limitedRequest)

	userInfo := RegisterUserRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 422, fieldErrorUser, err.Error())
		return
	}

	if err := h.Validate.Struct(userInfo); err != nil {
		h.RespondWithValidationErrors(w, err, "validation failed for register user")
		return
	}

	if userInfo.User.Email == "" || userInfo.User.Username == "" {
		h.RespondWithError(w, 422, fieldErrorUser, "user register attempt failed: no email or username")
		return
	}

	if userInfo.User.Password == "" {
		h.RespondWithError(w, 422, fieldErrorUser, "user register attempt failed: no password")
		return
	}

	userExists, err := h.DbQueries.DoesUsernameExist(r.Context(), userInfo.User.Username)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorUsername, err)
		return
	}
	if userExists {
		h.RespondWithError(w, 409, fieldErrorUsername, "username already taken")
		return
	}

	emailExists, err := h.DbQueries.DoesEmailExist(r.Context(), userInfo.User.Email)
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorEmail, err)
		return
	}
	if emailExists {
		h.RespondWithError(w, 409, fieldErrorEmail, "email already taken")
		return
	}

	hashedPW, err := auth.HashPassword(userInfo.User.Password)
	if err != nil {
		h.RespondWithError(w, 500, fieldErrorUser, fmt.Sprintf("could not hash password: %v", err))
		return
	}

	user, err := h.DbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Username:       userInfo.User.Username,
		Email:          userInfo.User.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		h.RespondWithDatabaseError(w, fieldErrorUser, err)
		return
	}

	token, err := h.Auth.MakeJWT(user.ID, JWTExpiresIn)
	if err != nil {
		h.RespondWithError(w, 201, "token", fmt.Sprintf("User %v registered, but could not create session: %v", user.ID, err))
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
		},
	}

	h.RespondWithJSON(w, 201, respBody)

}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {

	limitedRequest := http.MaxBytesReader(w, r.Body, 2048)
	decoder := json.NewDecoder(limitedRequest)

	userInfo := LoginUserRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 401, fieldErrorUser, err.Error())
		return
	}

	if err := h.Validate.Struct(userInfo); err != nil {
		h.RespondWithValidationErrors(w, err, "validation failed for login user")
		return
	}

	if userInfo.User.Email == "" {
		h.RespondWithError(w, 422, fieldErrorUser, "user login attempt failed: no email")
		return
	}

	user, err := h.DbQueries.GetUserByEmail(r.Context(), userInfo.User.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, fieldErrorUser, fmt.Sprintf("login attempt failed, no user found for email %v", userInfo.User.Email))
		} else {
			h.RespondWithDatabaseError(w, fieldErrorUser, err)
		}
		return
	}

	matching, err := auth.CheckPasswordHash(userInfo.User.Password, user.HashedPassword)
	if err != nil {
		h.RespondWithError(w, 401, fieldErrorUser, fmt.Sprintf("login attempt failed for user %v - %v", user.ID, err))
		return
	}

	if !matching {
		h.RespondWithError(w, 401, fieldErrorPassword, fmt.Sprintf("login attempt failed for user %v - wrong password", user.ID))
		return
	}

	token, err := h.Auth.MakeJWT(user.ID, JWTExpiresIn)
	if err != nil {
		h.RespondWithError(w, 500, fieldErrorUser, fmt.Sprintf("user %v logged in successfully, but could not create session: %v", user.ID, err))
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
		},
	}

	h.RespondWithJSON(w, 200, respBody)
}

func (h *Handler) CurrentUser(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(contextKeyUserID).(uuid.UUID)
	token := r.Context().Value(contextKeyToken).(string)

	user, err := h.DbQueries.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, fieldErrorUser, fmt.Sprintf("CurrentUser request failed, no user found for id %v", userID))
		} else {
			h.RespondWithDatabaseError(w, fieldErrorUser, err)
		}
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
		},
	}

	h.RespondWithJSON(w, 200, respBody)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(contextKeyUserID).(uuid.UUID)
	token := r.Context().Value(contextKeyToken).(string)

	decoder := json.NewDecoder(r.Body)

	userInfo := UpdateUserRequest{}

	if err := decoder.Decode(&userInfo); err != nil {
		h.RespondWithError(w, 401, fieldErrorUser, err.Error())
		return
	}

	if err := h.Validate.Struct(userInfo); err != nil {
		h.RespondWithValidationErrors(w, err, "validation failed for updating user")
		return
	}

	updateInfo := database.UpdateUserByIDParams{
		ID: userID,
	}
	if userInfo.User.Username != "" {
		updateInfo.Username = stringToNullString(userInfo.User.Username)
	}
	if userInfo.User.Email != "" {
		updateInfo.Email = stringToNullString(userInfo.User.Email)
	}
	if userInfo.User.Password != "" {
		if len(userInfo.User.Password) < minPasswordLength {
			h.RespondWithError(w, 422, fieldErrorPassword, "password too short")
		}

		hashPW, err := auth.HashPassword(userInfo.User.Password)
		if err != nil {
			h.RespondWithError(w, 500, fieldErrorUser, err.Error())
			return
		}
		updateInfo.HashedPassword = stringToNullString(hashPW)
	}

	if userInfo.User.Bio.Set {
		if userInfo.User.Bio.Value == nil || *userInfo.User.Bio.Value == "" {
			_, err := h.DbQueries.ClearUserBio(r.Context(), userID)
			if err != nil {
				h.RespondWithDatabaseError(w, fieldErrorUser, err)
				return
			}
		} else {
			updateInfo.Bio = pointerStringToNullString(userInfo.User.Bio.Value)
		}
	}

	if userInfo.User.Image.Set {
		if userInfo.User.Image.Value == nil || *userInfo.User.Image.Value == "" {
			_, err := h.DbQueries.ClearUserImage(r.Context(), userID)
			if err != nil {
				h.RespondWithDatabaseError(w, fieldErrorUser, err)
				return
			}
		} else {
			updateInfo.Image = pointerStringToNullString(userInfo.User.Image.Value)
		}
	}

	user, err := h.DbQueries.UpdateUserByID(r.Context(), updateInfo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.RespondWithError(w, 401, fieldErrorUser, fmt.Sprintf("UpdateUser request failed, no user found for id %v", userID))
		} else {
			h.RespondWithDatabaseError(w, fieldErrorUser, err)
		}
		return
	}

	respBody := RespondUser{
		User: User{
			Username: user.Username,
			Email:    user.Email,
			Token:    token,
			Bio:      nullStringToStringPointer(user.Bio),
			Image:    nullStringToStringPointer(user.Image),
		},
	}

	h.RespondWithJSON(w, 200, respBody)

}

// only for api testing
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(contextKeyUserID).(uuid.UUID)

	if err := h.DbQueries.DeleteUserByID(r.Context(), userID); err != nil {
		h.RespondWithDatabaseError(w, fieldErrorUser, err)
		return
	}

	w.WriteHeader(204)

}

func pointerStringToNullString(text *string) sql.NullString {
	if text == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *text, Valid: *text != ""}
}

func stringToNullString(text string) sql.NullString {
	return sql.NullString{
		String: text,
		Valid:  text != "",
	}
}

func nullStringToStringPointer(text sql.NullString) *string {
	if text.Valid {
		return &text.String
	}
	return nil
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	n.Value = &v
	return nil
}
