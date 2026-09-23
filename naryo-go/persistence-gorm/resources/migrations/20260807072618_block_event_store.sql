-- Create "solana_block_events" table
CREATE TABLE "public"."solana_block_events" (
  "id" uuid NOT NULL,
  "node_id" uuid NOT NULL,
  "slot" bigint NOT NULL,
  "blockhash" character varying(44) NOT NULL,
  "block_time" bigint NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_solana_block_events_blockhash" to table: "solana_block_events"
CREATE UNIQUE INDEX "idx_solana_block_events_blockhash" ON "public"."solana_block_events" ("blockhash");
-- Create index "idx_solana_block_events_node_slot" to table: "solana_block_events"
CREATE UNIQUE INDEX "idx_solana_block_events_node_slot" ON "public"."solana_block_events" ("node_id", "slot");
-- Modify "solana_transaction_events" table
ALTER TABLE "public"."solana_transaction_events" ADD CONSTRAINT "fk_solana_block_events_transactions" FOREIGN KEY ("node_id", "slot") REFERENCES "public"."solana_block_events" ("node_id", "slot") ON UPDATE NO ACTION ON DELETE CASCADE;
