-- Create "solana_transaction_events" table
CREATE TABLE "public"."solana_transaction_events" (
  "id" uuid NOT NULL,
  "node_id" uuid NOT NULL,
  "transaction_id" character varying(88) NOT NULL,
  "err" character varying(5000) NULL,
  "slot" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_solana_transaction_events_node_transaction" to table: "solana_transaction_events"
CREATE UNIQUE INDEX "idx_solana_transaction_events_node_transaction" ON "public"."solana_transaction_events" ("node_id", "transaction_id");
-- Create index "idx_solana_transaction_events_slot" to table: "solana_transaction_events"
CREATE INDEX "idx_solana_transaction_events_slot" ON "public"."solana_transaction_events" ("slot");
-- Create "solana_instructions" table
CREATE TABLE "public"."solana_instructions" (
  "id" uuid NOT NULL,
  "transaction_event_id" uuid NOT NULL,
  "instruction_index" bigint NOT NULL,
  "program_id" character varying(128) NOT NULL,
  "data" bytea NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_transaction_events_instructions" FOREIGN KEY ("transaction_event_id") REFERENCES "public"."solana_transaction_events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_instructions_transaction_event_instruction" to table: "solana_instructions"
CREATE UNIQUE INDEX "idx_solana_instructions_transaction_event_instruction" ON "public"."solana_instructions" ("transaction_event_id", "instruction_index");
-- Create "solana_instruction_accounts" table
CREATE TABLE "public"."solana_instruction_accounts" (
  "id" uuid NOT NULL,
  "instruction_id" uuid NOT NULL,
  "account_index" bigint NOT NULL,
  "address" character varying(64) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_instructions_accounts" FOREIGN KEY ("instruction_id") REFERENCES "public"."solana_instructions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_instruction_accounts_instruction_account" to table: "solana_instruction_accounts"
CREATE UNIQUE INDEX "idx_solana_instruction_accounts_instruction_account" ON "public"."solana_instruction_accounts" ("instruction_id", "account_index");
-- Create "solana_transaction_logs" table
CREATE TABLE "public"."solana_transaction_logs" (
  "id" uuid NOT NULL,
  "transaction_event_id" uuid NOT NULL,
  "log_index" bigint NOT NULL,
  "message" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_transaction_events_logs" FOREIGN KEY ("transaction_event_id") REFERENCES "public"."solana_transaction_events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_transaction_logs_transaction_event_log" to table: "solana_transaction_logs"
CREATE UNIQUE INDEX "idx_solana_transaction_logs_transaction_event_log" ON "public"."solana_transaction_logs" ("transaction_event_id", "log_index");
