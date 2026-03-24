-- +goose Up
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,
    type VARCHAR(64) NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    expires_at TIMESTAMP NOT NULL DEFAULT now()
);
