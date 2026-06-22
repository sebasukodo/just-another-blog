-- name: CreateTags :one
INSERT INTO tags (name)
VALUES(
    $1
)
ON CONFLICT(
    name
)
DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: CreateArticleTags :one
INSERT INTO article_tags (article_id, tag_id)
VALUES(
    $1,
    $2
)
RETURNING *;

-- name: DeleteArticleTagsByArticleID :exec
DELETE FROM article_tags
WHERE article_id = $1;

-- name: GetTags :many
SELECT name FROM tags;
