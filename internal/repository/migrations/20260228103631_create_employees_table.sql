-- +goose Up

CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    position VARCHAR(255) NOT NULL,
    hired_at DATE NULL,
    created_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS employees;
