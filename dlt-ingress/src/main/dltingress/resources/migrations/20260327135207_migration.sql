-- Create "event_stores" table
CREATE TABLE "event_stores" ("id" uuid NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "deleted_at" timestamptz NULL, "type" character varying(255) NOT NULL, "payload" jsonb NOT NULL, "trace_parent" character varying(55) NOT NULL, "publication_type" character varying(10) NOT NULL DEFAULT 'MEMORY', "is_cross" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- Create index "idx_event_stores_deleted_at" to table: "event_stores"
CREATE INDEX "idx_event_stores_deleted_at" ON "event_stores" ("deleted_at");
-- Create index "idx_event_stores_tx_hash_ordered" to table: "event_stores"
CREATE INDEX "idx_event_stores_tx_hash_ordered" ON "event_stores" (((payload ->> 'TransactionHash'::text))) WHERE (deleted_at IS NULL);
-- Create "event_consumers" table
CREATE TABLE "event_consumers" ("id" uuid NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "deleted_at" timestamptz NULL, "type" character varying(255) NOT NULL, "status" character varying(10) NOT NULL, "event_store_id" uuid NOT NULL, "is_cross" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"), CONSTRAINT "fk_event_stores_event_consumers" FOREIGN KEY ("event_store_id") REFERENCES "event_stores" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "idx_event_consumers_deleted_at" to table: "event_consumers"
CREATE INDEX "idx_event_consumers_deleted_at" ON "event_consumers" ("deleted_at");
