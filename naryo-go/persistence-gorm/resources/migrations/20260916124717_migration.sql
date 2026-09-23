-- Create "solana_latest_blocks" table
CREATE TABLE "public"."solana_latest_blocks" (
  "id" uuid NOT NULL,
  "node_id" uuid NOT NULL,
  "slot" bigint NOT NULL,
  "blockhash" character varying(44) NOT NULL,
  "block_time" bigint NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_solana_latest_block_node" to table: "solana_latest_blocks"
CREATE UNIQUE INDEX "idx_solana_latest_block_node" ON "public"."solana_latest_blocks" ("node_id");
