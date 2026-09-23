-- Create "svm_transactions" table
CREATE TABLE "svm_transactions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tx_id" character varying(255) NOT NULL,
  "network_id" character varying(255) NOT NULL,
  "network_url" character varying(255) NOT NULL,
  "dlt" character varying(20) NOT NULL,
  "fee_payer" character varying(255) NOT NULL,
  "recent_blockhash" character varying(100) NOT NULL,
  "serialized_transaction" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "svm_transactions_tx_id_key" UNIQUE ("tx_id")
);
-- Create index "idx_svm_transactions_deleted_at" to table: "svm_transactions"
CREATE INDEX "idx_svm_transactions_deleted_at" ON "svm_transactions" ("deleted_at");
