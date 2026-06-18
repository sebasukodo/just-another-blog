-- name: FollowUser :one
INSERT INTO user_follows(follower_id, following_id)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: UnfollowUser :one
DELETE FROM user_follows
WHERE follower_id = $1 AND following_id = $2
RETURNING follower_id;

-- name: IsFollowing :one
SELECT EXISTS (
    SELECT 1
    FROM user_follows
    WHERE follower_id = $1
      AND following_id = $2
);

-- name: GetFollowersOfAUser :many
SELECT u.*
FROM users u
JOIN user_follows as uf
ON u.id = uf.follower_id
WHERE uf.following_id = $1;

-- name: GetAllUsersAUserIsFollowing :many
SELECT u.*
FROM users u
JOIN user_follows as uf
ON u.id = uf.following_id
WHERE uf.follower_id = $1;
