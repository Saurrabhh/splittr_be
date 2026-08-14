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

-- Seed Data
INSERT INTO app_versions (id, min_supported_version, latest_version, force_update, ios_update_url, android_update_url, update_message) VALUES
(1, '1.0.0', '1.2.0', false, 'https://apps.apple.com/app/id123456789', 'https://play.google.com/store/apps/details?id=com.splittr.app', 'A new version of Splittr is available.')
ON CONFLICT (id) DO NOTHING;

INSERT INTO maintenance_status (id, in_maintenance, read_only_mode, message) VALUES
(1, false, false, 'Splittr is under scheduled maintenance.')
ON CONFLICT (id) DO NOTHING;

INSERT INTO system_limits (id, max_expense_amount, max_group_members, max_split_participants, max_receipt_size_mb, allowed_receipt_mime_types) VALUES
(1, 100000.00, 50, 50, 10, ARRAY['image/jpeg', 'image/png', 'application/pdf'])
ON CONFLICT (id) DO NOTHING;

INSERT INTO feature_flags (key, is_enabled, description) VALUES
('enableOcrReceiptScan', true, 'OCR receipt scanning'),
('enableSettlementReminders', true, 'Push notification reminders for settlements'),
('enableExportPdf', true, 'Export group summary as PDF'),
('enableGroupAnalytics', false, 'Group spending analytics charts')
ON CONFLICT (key) DO NOTHING;

INSERT INTO legal_configs (id, terms_of_service_url, privacy_policy_url, faq_url, support_email) VALUES
(1, 'https://splittr.app/terms', 'https://splittr.app/privacy', 'https://splittr.app/faq', 'support@splittr.app')
ON CONFLICT (id) DO NOTHING;

INSERT INTO expense_categories (id, name, icon_url, display_order) VALUES
('cat_food', 'Food & Dining', 'https://assets.splittr.app/categories/food.png', 1),
('cat_rent', 'Rent & Utilities', 'https://assets.splittr.app/categories/rent.png', 2),
('cat_travel', 'Travel & Transport', 'https://assets.splittr.app/categories/travel.png', 3),
('cat_entertainment', 'Entertainment', 'https://assets.splittr.app/categories/movie.png', 4),
('cat_shopping', 'Shopping', 'https://assets.splittr.app/categories/shopping.png', 5),
('cat_other', 'Other', 'https://assets.splittr.app/categories/other.png', 99)
ON CONFLICT (id) DO NOTHING;

INSERT INTO currencies (code, symbol, name, decimal_places, is_default) VALUES
('USD', '$', 'US Dollar', 2, TRUE),
('INR', '₹', 'Indian Rupee', 2, FALSE),
('EUR', '€', 'Euro', 2, FALSE),
('GBP', '£', 'British Pound', 2, FALSE)
ON CONFLICT (code) DO NOTHING;

INSERT INTO config_versions (id, version_hash) VALUES
(1, 'v1.0.0-initial')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS config_versions;
DROP TABLE IF EXISTS currencies;
DROP TABLE IF EXISTS expense_categories;
DROP TABLE IF EXISTS legal_configs;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS system_limits;
DROP TABLE IF EXISTS maintenance_status;
DROP TABLE IF EXISTS app_versions;
