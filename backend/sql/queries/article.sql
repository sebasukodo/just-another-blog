-- name: CreateArticle :one
INSERT INTO articles(author_id, slug, title, description, body)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: CreateTags :one
INSERT INTO tags (display_name, normalized_name)
VALUES(
    $1,
    $2
)
ON CONFLICT(
    normalized_name
)
DO UPDATE SET normalized_name = EXCLUDED.normalized_name
RETURNING *;

-- name: CreateArticleTags :one
INSERT INTO article_tags (article_id, tag_id)
VALUES(
    $1,
    $2
)
RETURNING *;
