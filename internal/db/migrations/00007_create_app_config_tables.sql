-- +goose Up
CREATE TABLE IF NOT EXISTS app_versions (
    id INT PRIMARY KEY DEFAULT 1,
    min_supported_version VARCHAR(20) NOT NULL,
    latest_version VARCHAR(20) NOT NULL,
    force_update BOOLEAN DEFAULT FALSE NOT NULL,
    ios_update_url TEXT NOT NULL,
    android_update_url TEXT NOT NULL,
    update_message TEXT NOT NULL,
    CONSTRAINT single_row_app_version CHECK (id = 1)
);

CREATE TABLE IF NOT EXISTS maintenance_status (
    id INT PRIMARY KEY DEFAULT 1,
    in_maintenance BOOLEAN DEFAULT FALSE NOT NULL,
    read_only_mode BOOLEAN DEFAULT FALSE NOT NULL,
    message TEXT NOT NULL,
    estimated_end_time TIMESTAMP WITH TIME ZONE,
    CONSTRAINT single_row_maintenance CHECK (id = 1)
);

CREATE TABLE IF NOT EXISTS system_limits (
    id INT PRIMARY KEY DEFAULT 1,
    max_expense_amount NUMERIC(12, 2) NOT NULL,
    max_group_members INT NOT NULL,
    max_split_participants INT NOT NULL,
    max_receipt_size_mb INT NOT NULL,
    allowed_receipt_mime_types TEXT[] NOT NULL,
    CONSTRAINT single_row_limits CHECK (id = 1)
);

CREATE TABLE IF NOT EXISTS feature_flags (
    key VARCHAR(100) PRIMARY KEY,
    is_enabled BOOLEAN DEFAULT FALSE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS legal_configs (
    id INT PRIMARY KEY DEFAULT 1,
    terms_of_service_url TEXT NOT NULL,
    privacy_policy_url TEXT NOT NULL,
    faq_url TEXT NOT NULL,
    support_email VARCHAR(255) NOT NULL,
    CONSTRAINT single_row_legal CHECK (id = 1)
);

CREATE TABLE IF NOT EXISTS expense_categories (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    icon_url TEXT NOT NULL,
    display_order INT DEFAULT 0 NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(10) PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    decimal_places INT DEFAULT 2 NOT NULL,
    is_default BOOLEAN DEFAULT FALSE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL
);

CREATE TABLE IF NOT EXISTS config_versions (
    id INT PRIMARY KEY DEFAULT 1,
    version_hash VARCHAR(64) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT single_row_config_version CHECK (id = 1)
);

-- +goose Down
DROP TABLE IF EXISTS config_versions;
DROP TABLE IF EXISTS currencies;
DROP TABLE IF EXISTS expense_categories;
DROP TABLE IF EXISTS legal_configs;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS system_limits;
DROP TABLE IF EXISTS maintenance_status;
DROP TABLE IF EXISTS app_versions;
