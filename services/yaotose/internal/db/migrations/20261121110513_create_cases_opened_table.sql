-- Create "cases_opened" table
CREATE TABLE "yaotose"."cases_opened" (
  "case_id" text NOT NULL,
  "kafka_topic" text NOT NULL,
  "kafka_case" text NOT NULL,
  "payload" bytea NOT NULL,
  "status" text NOT NULL,
  "created_at" bigint NOT NULL DEFAULT (EXTRACT(epoch FROM now()))::bigint,
  "published_at" bigint NULL,
  CONSTRAINT "cases_opened_case_id_pkey" PRIMARY KEY ("case_id")
);
