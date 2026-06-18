-- +goose Up
CREATE TABLE IF NOT EXISTS article_comments(
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    article_id BIGINT NOT NULL REFERENCES articles(id),
    author_id UUID NOT NULL REFERENCES users(id),
    body TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS article_comments;
