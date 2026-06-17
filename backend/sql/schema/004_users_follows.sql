-- +goose Up
CREATE TABLE IF NOT EXISTS user_follows(
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id <> following_id)
);

CREATE INDEX idx_user_follows_following_id
ON user_follows (following_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_follows_following_id;
DROP TABLE IF EXISTS user_follows;
