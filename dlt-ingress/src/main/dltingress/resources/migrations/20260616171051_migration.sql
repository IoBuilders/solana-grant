-- Create "failed_transactions" table
CREATE TABLE "failed_transactions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tx_id" character varying(255) NOT NULL,
  "network_id" character varying(255) NOT NULL,
  "status" character varying(20) NOT NULL,
  "error_details" character varying(5000) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "failed_transactions_tx_id_key" UNIQUE ("tx_id")
);
-- Create index "idx_failed_transactions_deleted_at" to table: "failed_transactions"
CREATE INDEX "idx_failed_transactions_deleted_at" ON "failed_transactions" ("deleted_at");
