-- Create index "idx_accounts_customer_id" to table: "accounts"
CREATE INDEX "idx_accounts_customer_id" ON "bank"."accounts" ("customer_id");
-- Create "devices" table
CREATE TABLE "bank"."devices" (
  "customer_id" uuid NOT NULL,
  "device_id" text NOT NULL,
  "platform" text NOT NULL,
  "os_version" text NULL,
  "app_version" text NULL,
  "first_seen_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_seen_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "bound_at" timestamptz NULL,
  "is_trusted" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("customer_id", "device_id"),
  CONSTRAINT "devices_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "bank"."customers" ("customer_id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "devices_app_version_check" CHECK ((app_version IS NULL) OR (length(TRIM(BOTH FROM app_version)) <= 32)),
  CONSTRAINT "devices_device_id_check" CHECK ((length(TRIM(BOTH FROM device_id)) >= 1) AND (length(TRIM(BOTH FROM device_id)) <= 128)),
  CONSTRAINT "devices_last_seen_at_check" CHECK (last_seen_at >= first_seen_at),
  CONSTRAINT "devices_os_version_check" CHECK ((os_version IS NULL) OR (length(TRIM(BOTH FROM os_version)) <= 32)),
  CONSTRAINT "devices_platform_check" CHECK (platform = ANY (ARRAY['IOS'::text, 'ANDROID'::text, 'WEB'::text]))
);
