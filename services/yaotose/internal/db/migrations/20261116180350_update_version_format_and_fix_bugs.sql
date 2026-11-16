-- Modify "blacklist" table
ALTER TABLE "yaotose"."blacklist" ALTER COLUMN "key_id" SET DEFAULT 'k1';
-- Modify "payer_payee" table
ALTER TABLE "yaotose"."payer_payee" ADD CONSTRAINT "payer_payee_total_amount_check" CHECK (total_amount > 0), ALTER COLUMN "min_amount" TYPE bigint, ALTER COLUMN "max_amount" TYPE bigint, ADD COLUMN "key_id" text NOT NULL DEFAULT 'k1', ADD COLUMN "total_amount" bigint NOT NULL DEFAULT 1;
