-- Create "sessions" table
CREATE TABLE "bank"."sessions" (
  "session_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "customer_id" uuid NOT NULL,
  "device_id" text NOT NULL,
  "channel" text NOT NULL,
  "started_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "ended_at" timestamptz NULL,
  "ip" inet NULL,
  "ip_country" character(2) NULL,
  "is_proxy" boolean NOT NULL DEFAULT false,
  "auth_method" text NULL,
  CONSTRAINT "sessions_session_id_pkey" PRIMARY KEY ("session_id"),
  CONSTRAINT "sessions_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "bank"."customers" ("customer_id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sessions_auth_method_check" CHECK ((auth_method IS NULL) OR (auth_method = ANY (ARRAY['PIN'::text, 'BIOMETRIC'::text, 'PASSWORD_2FA'::text]))),
  CONSTRAINT "sessions_channel_check" CHECK (channel = ANY (ARRAY['MOBILE_APP'::text, 'WEB'::text])),
  CONSTRAINT "sessions_device_id_check" CHECK ((length(TRIM(BOTH FROM device_id)) >= 1) AND (length(TRIM(BOTH FROM device_id)) <= 128)),
  CONSTRAINT "sessions_ended_at_check" CHECK ((ended_at IS NULL) OR (ended_at >= started_at)),
  CONSTRAINT "sessions_ip_country_check" CHECK ((ip_country IS NULL) OR (ip_country ~ '^[A-Z]{2}$'::text))
);
