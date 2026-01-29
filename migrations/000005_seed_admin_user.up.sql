-- Seed file: Create default admin user
-- Default credentials: username=admin, password=admin
-- Password hash for "admin" (BCrypt)
-- Generated using: bcrypt cost 10

INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
VALUES (
    'admin',
    'admin@todolist.local',
    '$2a$10$YgADynSCq4W5pVCpKwsNYesXoO9G65WcmyzTQ/DLEGUZBMGj1Ztbq', -- bcrypt hash for "admin"
    'admin',
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE updated_at = NOW();
