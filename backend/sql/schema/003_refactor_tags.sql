-- +goose Up
ALTER TABLE tags
RENAME COLUMN normalized_name TO name;

ALTER TABLE tags
DROP COLUMN display_name;

-- +goose Down
ALTER TABLE tags
ADD display_name TEXT NOT NULL DEFAULT '';

UPDATE tags
SET display_name = name;

ALTER TABLE tags
ALTER COLUMN display_name DROP DEFAULT;

ALTER TABLE tags
RENAME COLUMN name TO normalized_name;
