-- Create "beneficiaries" table
CREATE TABLE "bank"."beneficiaries" (
  "beneficiary_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NOT NULL,
  "name" text NOT NULL,
  "iban" text NOT NULL,
  "bic" character varying(11) NULL,
  "nickname" text NULL,
  "added_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "added_session_id" uuid NULL,
  "vop_result" text NULL,
  "status" text NOT NULL DEFAULT 'ACTIVE',
  "last_used_at" timestamptz NULL,
  CONSTRAINT "beneficiaries_beneficiary_id_pkey" PRIMARY KEY ("beneficiary_id"),
  CONSTRAINT "beneficiaries_customer_id_iban_key" UNIQUE ("customer_id", "iban"),
  CONSTRAINT "beneficiaries_added_session_id_fkey" FOREIGN KEY ("added_session_id") REFERENCES "bank"."sessions" ("session_id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "beneficiaries_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "bank"."customers" ("customer_id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "beneficiaries_bic_check" CHECK ((bic IS NULL) OR ((bic)::text ~ '^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$'::text)),
  CONSTRAINT "beneficiaries_iban_check" CHECK (iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$'::text),
  CONSTRAINT "beneficiaries_name_check" CHECK ((length(TRIM(BOTH FROM name)) >= 1) AND (length(TRIM(BOTH FROM name)) <= 140)),
  CONSTRAINT "beneficiaries_nickname_check" CHECK ((nickname IS NULL) OR ((length(TRIM(BOTH FROM nickname)) >= 1) AND (length(TRIM(BOTH FROM nickname)) <= 50))),
  CONSTRAINT "beneficiaries_status_check" CHECK (status = ANY (ARRAY['ACTIVE'::text, 'REMOVED'::text])),
  CONSTRAINT "beneficiaries_vop_result_check" CHECK ((vop_result IS NULL) OR (vop_result = ANY (ARRAY['MATCH'::text, 'CLOSE_MATCH'::text, 'NO_MATCH'::text, 'NOT_POSSIBLE'::text])))
);
