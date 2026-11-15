-- Create "transactions" table
CREATE TABLE "bank"."transactions" (
  "txn_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "account_id" uuid NOT NULL,
  "payment_id" uuid NULL,
  "booked_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "booking_date" date NOT NULL,
  "value_date" date NOT NULL,
  "direction" text NOT NULL,
  "amount_minor" bigint NOT NULL,
  "currency" character(3) NOT NULL,
  "balance_after_minor" bigint NOT NULL,
  "scheme" text NOT NULL,
  "counterparty_name" text NULL,
  "counterparty_iban" text NULL,
  "counterparty_bic" character varying(11) NULL,
  "remittance_info" text NULL,
  "end_to_end_id" text NULL,
  "channel" text NULL,
  "status" text NOT NULL DEFAULT 'BOOKED',
  CONSTRAINT "transactions_txn_id_pkey" PRIMARY KEY ("txn_id"),
  CONSTRAINT "transactions_account_id_fkey" FOREIGN KEY ("account_id") REFERENCES "bank"."accounts" ("account_id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "transactions_payment_id_fkey" FOREIGN KEY ("payment_id") REFERENCES "bank"."payment_orders" ("payment_id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transactions_amount_minor_check" CHECK (amount_minor > 0),
  CONSTRAINT "transactions_channel_check" CHECK ((channel IS NULL) OR (channel = ANY (ARRAY['MOBILE_APP'::text, 'WEB'::text, 'API'::text, 'BRANCH'::text, 'ATM'::text, 'POS'::text]))),
  CONSTRAINT "transactions_counterparty_bic_check" CHECK ((counterparty_bic IS NULL) OR ((counterparty_bic)::text ~ '^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$'::text)),
  CONSTRAINT "transactions_counterparty_iban_check" CHECK ((counterparty_iban IS NULL) OR (counterparty_iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'::text)),
  CONSTRAINT "transactions_counterparty_name_check" CHECK ((counterparty_name IS NULL) OR ((length(TRIM(BOTH FROM counterparty_name)) >= 1) AND (length(TRIM(BOTH FROM counterparty_name)) <= 140))),
  CONSTRAINT "transactions_currency_check" CHECK (currency ~ '^[A-Z]{3}$'::text),
  CONSTRAINT "transactions_direction_check" CHECK (direction = ANY (ARRAY['DEBIT'::text, 'CREDIT'::text])),
  CONSTRAINT "transactions_end_to_end_id_check" CHECK ((end_to_end_id IS NULL) OR ((length(TRIM(BOTH FROM end_to_end_id)) >= 1) AND (length(TRIM(BOTH FROM end_to_end_id)) <= 35))),
  CONSTRAINT "transactions_remittance_info_check" CHECK ((remittance_info IS NULL) OR (length(remittance_info) <= 140)),
  CONSTRAINT "transactions_scheme_check" CHECK (scheme = ANY (ARRAY['SEPA_INSTANT'::text, 'SEPA_CT'::text, 'SEPA_DD'::text, 'CARD'::text, 'ATM'::text, 'FEE'::text, 'INTEREST'::text, 'INTERNAL'::text])),
  CONSTRAINT "transactions_status_check" CHECK (status = ANY (ARRAY['PENDING'::text, 'BOOKED'::text, 'REVERSED'::text]))
);
-- Create index "idx_transactions_account_id_booked_at" to table: "transactions"
CREATE INDEX "idx_transactions_account_id_booked_at" ON "bank"."transactions" ("account_id", "booked_at" DESC);
-- Create index "idx_transactions_counterparty_iban_booked_at" to table: "transactions"
CREATE INDEX "idx_transactions_counterparty_iban_booked_at" ON "bank"."transactions" ("counterparty_iban", "booked_at" DESC);
