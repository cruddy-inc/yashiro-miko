-- Create "profile_changes" table
CREATE TABLE "bank"."profile_changes" (
  "change_id" bigserial NOT NULL,
  "customer_id" uuid NOT NULL,
  "field" text NOT NULL,
  "old_value_hash" text NULL,
  "new_value_hash" text NULL,
  "changed_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "session_id" uuid NULL,
  CONSTRAINT "profile_changes_change_id_pkey" PRIMARY KEY ("change_id"),
  CONSTRAINT "profile_changes_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "bank"."customers" ("customer_id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "profile_changes_session_id_fkey" FOREIGN KEY ("session_id") REFERENCES "bank"."sessions" ("session_id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "profile_changes_field_check" CHECK (field = ANY (ARRAY['PHONE'::text, 'EMAIL'::text, 'ADDRESS'::text, 'LIMIT'::text])),
  CONSTRAINT "profile_changes_new_value_hash_check" CHECK ((new_value_hash IS NULL) OR (length(TRIM(BOTH FROM new_value_hash)) >= 32)),
  CONSTRAINT "profile_changes_old_value_hash_check" CHECK ((old_value_hash IS NULL) OR (length(TRIM(BOTH FROM old_value_hash)) >= 32))
);
-- Create index "idx_profile_changes_customer_id_changed_at" to table: "profile_changes"
CREATE INDEX "idx_profile_changes_customer_id_changed_at" ON "bank"."profile_changes" ("customer_id", "changed_at" DESC);
