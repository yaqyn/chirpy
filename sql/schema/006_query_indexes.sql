-- +goose Up
CREATE INDEX chirps_created_at_id_idx ON chirps(created_at, id);
CREATE INDEX chirps_user_created_at_id_idx ON chirps(user_id, created_at, id);
CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens(user_id);

-- +goose Down
DROP INDEX refresh_tokens_user_id_idx;
DROP INDEX chirps_user_created_at_id_idx;
DROP INDEX chirps_created_at_id_idx;
