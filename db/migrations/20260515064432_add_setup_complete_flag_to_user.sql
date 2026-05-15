-- +goose Up
ALTER TABLE users
ADD COLUMN setup_complete BOOLEAN DEFAULT FALSE;

-- +goose Down
ALTER TABLE users 
DROP COLUMN setup_complete;
