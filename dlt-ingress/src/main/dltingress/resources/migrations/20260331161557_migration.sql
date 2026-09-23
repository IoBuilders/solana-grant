-- Create "custody_keys" table
CREATE TABLE "custody_keys" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "key_type" character varying(30) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'ACTIVE',
  "dlt_account_id" character varying(255) NOT NULL,
  "dlt" character varying(20) NOT NULL,
  "external_id" character varying(255) NOT NULL,
  "custody_provider" character varying(30) NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_custody_keys_deleted_at" to table: "custody_keys"
CREATE INDEX "idx_custody_keys_deleted_at" ON "custody_keys" ("deleted_at");
-- Create index "idx_custody_keys_dlt_account_id" to table: "custody_keys"
CREATE UNIQUE INDEX "idx_custody_keys_dlt_account_id" ON "custody_keys" ("dlt_account_id");
