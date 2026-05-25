-- name: CreateUser :one
INSERT INTO users (username, email, hashed_password, bio, image)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
