CREATE TABLE "service_accounts" (
  "id"             BIGSERIAL PRIMARY KEY,
  "name"           VARCHAR(100) NOT NULL,
  "description"    VARCHAR(255),
  "api_key_prefix" VARCHAR(16) NOT NULL,
  "api_key_hash"   VARCHAR(255) NOT NULL,
  "revoked"        BOOLEAN NOT NULL DEFAULT false,
  "last_used_at"   TIMESTAMP,
  "created_at"     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX ON "service_accounts" ("api_key_prefix");
