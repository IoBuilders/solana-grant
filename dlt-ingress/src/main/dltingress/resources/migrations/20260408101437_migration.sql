-- Create "ethereum_transactions" table
CREATE TABLE "ethereum_transactions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tx_id" character varying(255) NOT NULL,
  "network_id" character varying(255) NOT NULL,
  "network_url" character varying(255) NOT NULL,
  "dlt" character varying(20) NOT NULL,
  "from_address" character varying(255) NOT NULL,
  "to_address" character varying(255) NOT NULL,
  "nonce" character varying(100) NOT NULL,
  "value" character varying(100) NOT NULL,
  "transaction_type" bigint NOT NULL,
  "gas_limit" character varying(100) NOT NULL,
  "gas_price" character varying(100) NULL,
  "max_priority_fee_per_gas" character varying(100) NULL,
  "max_fee_per_gas" character varying(100) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ethereum_transactions_tx_id_key" UNIQUE ("tx_id")
);
-- Create index "idx_ethereum_transactions_deleted_at" to table: "ethereum_transactions"
CREATE INDEX "idx_ethereum_transactions_deleted_at" ON "ethereum_transactions" ("deleted_at");
