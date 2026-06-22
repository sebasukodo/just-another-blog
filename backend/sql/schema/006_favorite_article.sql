-- +goose Up
CREATE TABLE IF NOT EXISTS article_favorites(
    article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (article_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS article_favorites;
