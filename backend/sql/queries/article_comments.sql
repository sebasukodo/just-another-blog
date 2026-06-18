-- name: AddCommentToArticle :one
INSERT INTO article_comments(article_id, author_id, body)
VALUES(
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetCommentsFromArticle :many
SELECT * FROM article_comments
WHERE article_id = $1;
