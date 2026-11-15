-- Create "payment_orders" table
CREATE TABLE "bank"."payment_orders" (
  "payment_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "debtor_account_id" uuid NOT NULL,
  "creditor_name" text NOT NULL,
  "creditor_iban" text NOT NULL,
  "creditor_bic" character varying(11) NULL,
  "amount_minor" bigint NOT NULL,
  "currency" character(3) NOT NULL,
  "scheme" text NOT NULL,
  "end_to_end_id" text NOT NULL,
  "remittance_info" text NULL,
  "session_id" uuid NULL,
  "requested_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "status" text NOT NULL,
  "status_reason" text NULL,
  "executed_at" timestamptz NULL,
  CONSTRAINT "payment_orders_payment_id_pkey" PRIMARY KEY ("payment_id"),
  CONSTRAINT "payment_orders_debtor_account_id_fkey" FOREIGN KEY ("debtor_account_id") REFERENCES "bank"."accounts" ("account_id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "payment_orders_session_id_fkey" FOREIGN KEY ("session_id") REFERENCES "bank"."sessions" ("session_id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "payment_orders_amount_minor_check" CHECK (amount_minor > 0),
  CONSTRAINT "payment_orders_creditor_bic_check" CHECK ((creditor_bic IS NULL) OR ((creditor_bic)::text ~ '^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$'::text)),
  CONSTRAINT "payment_orders_creditor_iban_check" CHECK (creditor_iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'::text),
  CONSTRAINT "payment_orders_creditor_name_check" CHECK ((length(TRIM(BOTH FROM creditor_name)) >= 1) AND (length(TRIM(BOTH FROM creditor_name)) <= 140)),
  CONSTRAINT "payment_orders_currency_check" CHECK (currency ~ '^[A-Z]{3}$'::text),
  CONSTRAINT "payment_orders_end_to_end_id_check" CHECK ((length(TRIM(BOTH FROM end_to_end_id)) >= 1) AND (length(TRIM(BOTH FROM end_to_end_id)) <= 35)),
  CONSTRAINT "payment_orders_executed_at_check" CHECK ((executed_at IS NULL) OR (executed_at >= requested_at)),
  CONSTRAINT "payment_orders_remittance_info_check" CHECK ((remittance_info IS NULL) OR (length(remittance_info) <= 140)),
  CONSTRAINT "payment_orders_scheme_check" CHECK (scheme = ANY (ARRAY['SEPA_INSTANT'::text, 'SEPA_CT'::text, 'INTERNAL'::text])),
  CONSTRAINT "payment_orders_status_check" CHECK (status = ANY (ARRAY['RECEIVED'::text, 'SCREENING'::text, 'HELD'::text, 'EXECUTED'::text, 'REJECTED'::text, 'RETURNED'::text])),
  CONSTRAINT "payment_orders_status_reason_check" CHECK ((status_reason IS NULL) OR (length(TRIM(BOTH FROM status_reason)) <= 255))
);
-- Create index "idx_payment_orders_debtor_account_id_requested_at" to table: "payment_orders"
CREATE INDEX "idx_payment_orders_debtor_account_id_requested_at" ON "bank"."payment_orders" ("debtor_account_id", "requested_at" DESC);
