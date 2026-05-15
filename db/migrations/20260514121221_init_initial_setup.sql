-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- 1. The Financials & Rules Profile
CREATE TABLE user_profiles (
    user_id VARCHAR(255) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    monthly_salary NUMERIC(12, 2) NOT NULL,
    yearly_hike_percentage NUMERIC(5, 2) NOT NULL,
    needs_percentage INT NOT NULL,
    wants_percentage INT NOT NULL,
    savings_percentage INT NOT NULL,
    setup_completed BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Categories
CREATE TABLE categories (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('Need', 'Want', 'Saving')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Payment Methods
CREATE TABLE payment_methods (
    id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method_name VARCHAR(100) NOT NULL CHECK (method_name IN ('Cash', 'Debit Card', 'Credit Card', 'Upi', 'Bank')),
    weekly_limit NUMERIC(12, 2) NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
-- Notice how we DROP them in the exact reverse order they were created to prevent Foreign Key errors!

DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS user_profiles;