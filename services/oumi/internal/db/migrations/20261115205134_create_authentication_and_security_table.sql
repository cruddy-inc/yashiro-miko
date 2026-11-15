-- Create index "idx_sessions_customer_id_started_at" to table: "sessions"
CREATE INDEX "idx_sessions_customer_id_started_at" ON "bank"."sessions" ("customer_id", "started_at" DESC);
-- Create "auth_events" table
CREATE TABLE "bank"."auth_events" (
  "event_id" bigserial NOT NULL,
  "customer_id" uuid NOT NULL,
  "device_id" text NULL,
  "occurred_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "event_type" text NOT NULL,
  "ip" inet NULL,
  "ip_country" character(2) NULL,
  CONSTRAINT "auth_events_event_id_pkey" PRIMARY KEY ("event_id"),
  CONSTRAINT "auth_events_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "bank"."customers" ("customer_id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "auth_events_device_id_check" CHECK ((device_id IS NULL) OR ((length(TRIM(BOTH FROM device_id)) >= 1) AND (length(TRIM(BOTH FROM device_id)) <= 128))),
  CONSTRAINT "auth_events_event_type_check" CHECK (event_type = ANY (ARRAY['LOGIN_SUCCESS'::text, 'LOGIN_FAILED'::text, 'MFA_FAILED'::text, 'PASSWORD_RESET'::text, 'DEVICE_BOUND'::text])),
  CONSTRAINT "auth_events_ip_country_check" CHECK ((ip_country IS NULL) OR (ip_country ~ '^[A-Z]{2}$'::text))
);
-- Create index "idx_auth_events_customer_id_occurred_at" to table: "auth_events"
CREATE INDEX "idx_auth_events_customer_id_occurred_at" ON "bank"."auth_events" ("customer_id", "occurred_at" DESC);
