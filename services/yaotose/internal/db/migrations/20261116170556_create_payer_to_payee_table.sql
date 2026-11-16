-- Create "payer_payee" table
CREATE TABLE "yaotose"."payer_payee" (
  "tenant_id" text NOT NULL,
  "payer_hash" text NOT NULL,
  "payee_hash" text NOT NULL,
  "n_payments" integer NOT NULL DEFAULT 1,
  "min_amount" integer NOT NULL DEFAULT 0,
  "max_amount" integer NOT NULL DEFAULT 0,
  "first_transaction" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_transaction" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("tenant_id", "payer_hash", "payee_hash"),
  CONSTRAINT "payer_payee_max_amount_check" CHECK (max_amount > 0),
  CONSTRAINT "payer_payee_min_amount_check" CHECK (min_amount > 0),
  CONSTRAINT "payer_payee_n_payments_check" CHECK (n_payments > 0),
  CONSTRAINT "payer_payee_payee_hash_check" CHECK (payee_hash ~ '^[0-9a-f]{64}'::text),
  CONSTRAINT "payer_payee_payer_hash_check" CHECK (payer_hash ~ '^[0-9a-f]{64}'::text),
  CONSTRAINT "payer_payee_tenant_id_check" CHECK ((length(TRIM(BOTH FROM tenant_id)) >= 1) AND (length(TRIM(BOTH FROM tenant_id)) <= 64))
);
