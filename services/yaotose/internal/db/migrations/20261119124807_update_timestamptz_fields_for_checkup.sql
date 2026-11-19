-- Create index "idx_blacklist_expires_at" to table: "blacklist"
CREATE INDEX "idx_blacklist_expires_at" ON "yaotose"."blacklist" ("expires_at" DESC);
-- Modify "payer_payee" table
ALTER TABLE "yaotose"."payer_payee" ALTER COLUMN "first_transaction" TYPE bigint, ALTER COLUMN "first_transaction" SET DEFAULT (EXTRACT(epoch FROM now()))::bigint, ALTER COLUMN "last_transaction" TYPE bigint, ALTER COLUMN "last_transaction" SET DEFAULT (EXTRACT(epoch FROM now()))::bigint;
