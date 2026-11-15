-- Add new schema named "bank"
CREATE SCHEMA "bank";
-- Create "customers" table
CREATE TABLE "bank"."customers" (
  "customer_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_type" text NOT NULL,
  "full_name" text NOT NULL,
  "date_of_birth" date NULL,
  "country" character(2) NOT NULL,
  "email" text NULL,
  "phone_e164" text NULL,
  "segment" text NOT NULL DEFAULT 'MASS',
  "kyc_status" text NOT NULL,
  "kyc_risk_rating" text NOT NULL DEFAULT 'LOW',
  "is_pep" boolean NOT NULL DEFAULT false,
  "occupation" text NULL,
  "expected_monthly_turnover_minor" bigint NULL,
  "onboarded_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "status" text NOT NULL DEFAULT 'ACTIVE',
  CONSTRAINT "customers_customer_id_pkey" PRIMARY KEY ("customer_id"),
  CONSTRAINT "customers_email_key" UNIQUE ("email"),
  CONSTRAINT "customers_country_check" CHECK (country ~ '^[A-Z]{2}$'::text),
  CONSTRAINT "customers_customer_type_check" CHECK (customer_type = ANY (ARRAY['RETAIL'::text, 'SME'::text])),
  CONSTRAINT "customers_email_check" CHECK ((email IS NULL) OR (((length(TRIM(BOTH FROM email)) >= 6) AND (length(TRIM(BOTH FROM email)) <= 254)) AND (email ~* '^[a-z0-9._+-]{1,64}@([a-z0-9-]{1,63}\.)+([a-z0-9]{2,18})$'::text))),
  CONSTRAINT "customers_expected_monthly_turnover_minor_check" CHECK ((expected_monthly_turnover_minor IS NULL) OR (expected_monthly_turnover_minor >= 0)),
  CONSTRAINT "customers_full_name_check" CHECK ((length(TRIM(BOTH FROM full_name)) >= 2) AND (length(TRIM(BOTH FROM full_name)) <= 120)),
  CONSTRAINT "customers_kyc_risk_rating_check" CHECK (kyc_risk_rating = ANY (ARRAY['LOW'::text, 'MEDIUM'::text, 'HIGH'::text])),
  CONSTRAINT "customers_kyc_status_check" CHECK (kyc_status = ANY (ARRAY['PENDING'::text, 'VERIFIED'::text, 'EXPIRED'::text])),
  CONSTRAINT "customers_occupation_check" CHECK ((occupation IS NULL) OR ((length(TRIM(BOTH FROM occupation)) >= 2) AND (length(TRIM(BOTH FROM occupation)) <= 100))),
  CONSTRAINT "customers_phone_e164_check" CHECK ((phone_e164 IS NULL) OR (phone_e164 ~ '^\+[1-9][0-9]{6,14}$'::text)),
  CONSTRAINT "customers_segment_check" CHECK (segment = ANY (ARRAY['MASS'::text, 'AFFLUENT'::text, 'SME'::text])),
  CONSTRAINT "customers_status_check" CHECK (status = ANY (ARRAY['ACTIVE'::text, 'BLOCKED'::text, 'CLOSED'::text]))
);
