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

-- name: UpdateUserByID :one
UPDATE users
SET username = COALESCE(sqlc.narg('username'), username),
    email = COALESCE(sqlc.narg('email'), email),
    hashed_password = COALESCE(sqlc.narg('hashed_password'), hashed_password),
    bio = COALESCE(sqlc.narg('bio'), bio),
    image = COALESCE(sqlc.narg('image'), image),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ClearUserBio :one
UPDATE users
SET bio = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ClearUserImage :one
UPDATE users
SET image = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = $1;
