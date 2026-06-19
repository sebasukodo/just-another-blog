-- name: FavoriteArticle :one
INSERT INTO article_favorites(article_id, user_id)
VALUES(
    $1,
    $2
)
RETURNING *;

-- name: CountFavorites :one
SELECT COUNT(*)
FROM article_favorites
WHERE article_id = $1;

-- name: UnfavoriteArticle :exec
DELETE FROM article_favorites
WHERE article_id = $1 AND user_id = $2;
