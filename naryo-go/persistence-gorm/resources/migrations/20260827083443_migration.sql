-- Modify "solana_transaction_events" table
ALTER TABLE "public"."solana_transaction_events" DROP CONSTRAINT "fk_solana_block_events_transactions", ADD COLUMN "block_event_id" uuid NOT NULL, ADD CONSTRAINT "fk_solana_block_events_transactions" FOREIGN KEY ("block_event_id") REFERENCES "public"."solana_block_events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create "solana_transaction_accounts" table
CREATE TABLE "public"."solana_transaction_accounts" (
  "id" uuid NOT NULL,
  "transaction_event_id" uuid NOT NULL,
  "account_index" bigint NOT NULL,
  "address" character varying(64) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_transaction_events_accounts" FOREIGN KEY ("transaction_event_id") REFERENCES "public"."solana_transaction_events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_solana_transaction_accounts_transaction_event_account" to table: "solana_transaction_accounts"
CREATE UNIQUE INDEX "idx_solana_transaction_accounts_transaction_event_account" ON "public"."solana_transaction_accounts" ("transaction_event_id", "account_index");
