-- Drop triggers and functions
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Drop dependent tables first (foreign key dependencies)
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

-- Drop extensions (optional, only if unused elsewhere)
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
