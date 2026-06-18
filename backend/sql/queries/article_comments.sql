-- name: AddCommentToArticle :one
INSERT INTO article_comments(article_id, author_id, body)
VALUES(
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetCommentsFromArticle :many
SELECT c.id, c.article_id, c.body, c.author_id, c.created_at, c.updated_at, a.username, a.bio, a.image, EXISTS (
  SELECT 1
  FROM user_follows uf
  WHERE uf.follower_id = $2 AND uf.following_id = c.author_id
) AS author_is_followed
FROM article_comments c
JOIN users a
ON a.id = c.author_id
WHERE c.article_id = $1;
