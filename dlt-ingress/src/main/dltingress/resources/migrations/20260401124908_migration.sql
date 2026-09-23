-- Create "nonces" table
CREATE TABLE "nonces" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "dlt_account_id" character varying(255) NOT NULL,
  "network_id" character varying(255) NOT NULL,
  "value" character varying(100) NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_nonce_dltaccount_network" to table: "nonces"
CREATE UNIQUE INDEX "idx_nonce_dltaccount_network" ON "nonces" ("dlt_account_id", "network_id");
-- Create index "idx_nonces_deleted_at" to table: "nonces"
CREATE INDEX "idx_nonces_deleted_at" ON "nonces" ("deleted_at");
