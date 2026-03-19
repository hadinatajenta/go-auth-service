-- Align user_sessions schema with the application model.
-- The original migration created a single `token` column, but the
-- application uses separate access_token and refresh_token columns.
-- This migration applies the correct structure.

ALTER TABLE "user_sessions"
  ADD COLUMN IF NOT EXISTS "access_token"  VARCHAR(512),
  ADD COLUMN IF NOT EXISTS "refresh_token" VARCHAR(512);

-- Ensure ip_address, user_agent, created_at are present (idempotent)
ALTER TABLE "user_sessions"
  ADD COLUMN IF NOT EXISTS "ip_address"  VARCHAR(45),
  ADD COLUMN IF NOT EXISTS "user_agent"  VARCHAR(255),
  ADD COLUMN IF NOT EXISTS "created_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
