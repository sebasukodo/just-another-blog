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

-- name: GetArticleIDBySlug :one
SELECT id FROM articles WHERE slug = $1;

-- name: GetArticleTagsByArticleID :many
SELECT t.name FROM tags t
JOIN article_tags at ON t.id = at.tag_id
JOIN articles a ON at.article_id = a.id
WHERE a.id = $1;

-- name: UpdateArticleBySlug :one
UPDATE articles
SET title = COALESCE(sqlc.narg('title'), title),
    body = COALESCE(sqlc.narg('body'), body),
    description = COALESCE(sqlc.narg('description'), description),
    slug = COALESCE(sqlc.narg('new_slug'), slug),
    updated_at = NOW()
WHERE slug = sqlc.arg('slug')
RETURNING *;

-- name: DeleteArticleById :exec
DELETE FROM articles
WHERE id = $1;

-- name: FeedArticles :many
SELECT a.*, u.username, u.bio, u.image, array_agg(t.name ORDER BY t.name)::text[] AS tags
FROM articles a
JOIN users u
ON a.author_id = u.id
JOIN user_follows uf
ON uf.follower_id = $1 AND uf.following_id = a.author_id
LEFT JOIN article_tags at
ON at.article_id = a.id
LEFT JOIN tags t
ON t.id = at.tag_id
GROUP BY a.id, u.id
LIMIT $2
OFFSET $3;
