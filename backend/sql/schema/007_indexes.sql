-- +goose Up
CREATE INDEX IF NOT EXISTS idx_article_comments_article_id
ON article_comments (article_id);

CREATE INDEX IF NOT EXISTS idx_articles_author_id
ON articles (author_id);

CREATE INDEX IF NOT EXISTS idx_articles_created_at
ON articles (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_article_tags_tag_id
ON article_tags (tag_id);

CREATE INDEX IF NOT EXISTS idx_article_favorites_user_id
ON article_favorites (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_article_favorites_user_id;
DROP INDEX IF EXISTS idx_article_tags_tag_id;
DROP INDEX IF EXISTS idx_articles_created_at;
DROP INDEX IF EXISTS idx_articles_author_id;
DROP INDEX IF EXISTS idx_article_comments_article_id;
