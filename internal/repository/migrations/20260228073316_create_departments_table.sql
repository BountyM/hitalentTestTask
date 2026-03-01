-- +goose Up
CREATE TABLE departments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    parent_id BIGINT REFERENCES departments(id) ON DELETE SET NULL,
    created_at TIMESTAMP,
    UNIQUE (parent_id, name)
);

-- +goose Down
DROP TABLE IF EXISTS departments;
