-- +goose Up
-- +goose StatementBegin

CREATE TABLE expenses (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    category_id VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    payment_mode VARCHAR(100) NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Foreign Keys
    CONSTRAINT fk_user
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE,

    CONSTRAINT fk_category
        FOREIGN KEY (category_id) 
        REFERENCES categories(id) 
        ON DELETE RESTRICT
);

-- SDE3 Performance Tuning: Adding Indexes for fast data retrieval
-- 1. Index for the Dashboard/Weekly calculation (User + Date filtering)
CREATE INDEX idx_expenses_user_date ON expenses(user_id, date);

-- 2. Index for filtering expenses by category
CREATE INDEX idx_expenses_category ON expenses(category_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS expenses;

-- +goose StatementEnd