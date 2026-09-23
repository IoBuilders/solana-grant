-- Modify "solana_transaction_events" table
ALTER TABLE "public"."solana_transaction_events" DROP COLUMN "block_event_id";
-- Create "solana_block_transactions" table
CREATE TABLE "public"."solana_block_transactions" (
  "id" uuid NOT NULL,
  "block_event_id" uuid NOT NULL,
  "node_id" uuid NOT NULL,
  "transaction_id" character varying(88) NOT NULL,
  "err" character varying(5000) NULL,
  "slot" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_block_events_transactions" FOREIGN KEY ("block_event_id") REFERENCES "public"."solana_block_events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_block_transactions_node_transaction" to table: "solana_block_transactions"
CREATE UNIQUE INDEX "idx_solana_block_transactions_node_transaction" ON "public"."solana_block_transactions" ("node_id", "transaction_id");
-- Create index "idx_solana_block_transactions_slot" to table: "solana_block_transactions"
CREATE INDEX "idx_solana_block_transactions_slot" ON "public"."solana_block_transactions" ("slot");
-- Create "solana_block_instructions" table
CREATE TABLE "public"."solana_block_instructions" (
  "id" uuid NOT NULL,
  "block_transaction_id" uuid NOT NULL,
  "instruction_index" bigint NOT NULL,
  "program_id" character varying(128) NOT NULL,
  "data" bytea NULL,
  "inner" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_block_transactions_instructions" FOREIGN KEY ("block_transaction_id") REFERENCES "public"."solana_block_transactions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_block_instructions_block_transaction_instruction" to table: "solana_block_instructions"
CREATE UNIQUE INDEX "idx_solana_block_instructions_block_transaction_instruction" ON "public"."solana_block_instructions" ("block_transaction_id", "instruction_index");
-- Create "solana_block_instruction_accounts" table
CREATE TABLE "public"."solana_block_instruction_accounts" (
  "id" uuid NOT NULL,
  "block_instruction_id" uuid NOT NULL,
  "account_index" bigint NOT NULL,
  "address" character varying(64) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_block_instructions_accounts" FOREIGN KEY ("block_instruction_id") REFERENCES "public"."solana_block_instructions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_block_instruction_accounts_block_instruction_account" to table: "solana_block_instruction_accounts"
CREATE UNIQUE INDEX "idx_solana_block_instruction_accounts_block_instruction_account" ON "public"."solana_block_instruction_accounts" ("block_instruction_id", "account_index");
-- Create "solana_block_transaction_accounts" table
CREATE TABLE "public"."solana_block_transaction_accounts" (
  "id" uuid NOT NULL,
  "block_transaction_id" uuid NOT NULL,
  "account_index" bigint NOT NULL,
  "address" character varying(64) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_block_transactions_accounts" FOREIGN KEY ("block_transaction_id") REFERENCES "public"."solana_block_transactions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_block_transaction_accounts_block_transaction_account" to table: "solana_block_transaction_accounts"
CREATE UNIQUE INDEX "idx_solana_block_transaction_accounts_block_transaction_account" ON "public"."solana_block_transaction_accounts" ("block_transaction_id", "account_index");
-- Create "solana_block_transaction_logs" table
CREATE TABLE "public"."solana_block_transaction_logs" (
  "id" uuid NOT NULL,
  "block_transaction_id" uuid NOT NULL,
  "log_index" bigint NOT NULL,
  "message" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_block_transactions_logs" FOREIGN KEY ("block_transaction_id") REFERENCES "public"."solana_block_transactions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_block_transaction_logs_block_transaction_log" to table: "solana_block_transaction_logs"
CREATE UNIQUE INDEX "idx_solana_block_transaction_logs_block_transaction_log" ON "public"."solana_block_transaction_logs" ("block_transaction_id", "log_index");
