-- +goose Up
ALTER TABLE user_profiles
ADD COLUMN xx_weekly_limit NUMERIC(12, 2) NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE user_profiles
DROP COLUMN xx_weekly_limit;
