-- +goose Up
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
DELETE FROM config_versions WHERE id = 1;
DELETE FROM currencies WHERE code IN ('USD', 'INR', 'EUR', 'GBP');
DELETE FROM expense_categories WHERE id IN ('cat_food', 'cat_rent', 'cat_travel', 'cat_entertainment', 'cat_shopping', 'cat_other');
DELETE FROM legal_configs WHERE id = 1;
DELETE FROM feature_flags WHERE key IN ('enableOcrReceiptScan', 'enableSettlementReminders', 'enableExportPdf', 'enableGroupAnalytics');
DELETE FROM system_limits WHERE id = 1;
DELETE FROM maintenance_status WHERE id = 1;
DELETE FROM app_versions WHERE id = 1;
