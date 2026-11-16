-- Add new schema named "yaotose"
CREATE SCHEMA "yaotose";
-- Create "blacklist" table
CREATE TABLE "yaotose"."blacklist" (
  "tenant_id" text NOT NULL,
  "iban_hash" text NOT NULL,
  "key_id" text NOT NULL DEFAULT '1',
  "reason" text[] NOT NULL,
  "ai_message" text NOT NULL,
  "blacklisted_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("tenant_id", "iban_hash"),
  CONSTRAINT "blacklist_ai_message_check" CHECK (length(TRIM(BOTH FROM ai_message)) > 0),
  CONSTRAINT "blacklist_expires_at_check" CHECK ((expires_at IS NULL) OR (expires_at > blacklisted_at)),
  CONSTRAINT "blacklist_iban_hash_check" CHECK (iban_hash ~ '^[0-9a-f]{64}$'::text),
  CONSTRAINT "blacklist_key_id_check" CHECK ((length(TRIM(BOTH FROM key_id)) >= 1) AND (length(TRIM(BOTH FROM key_id)) <= 32)),
  CONSTRAINT "blacklist_tenant_id_check" CHECK ((length(TRIM(BOTH FROM tenant_id)) >= 1) AND (length(TRIM(BOTH FROM tenant_id)) <= 64))
);
