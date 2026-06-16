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

-- name: GetArticleBySlug :one
SELECT * FROM articles WHERE slug = $1;

-- name: GetArticleTagsByArticleID :many
SELECT t.name FROM tags t
JOIN article_tags at ON t.id = at.tag_id
JOIN articles a ON at.article_id = a.id
WHERE a.id = $1;
