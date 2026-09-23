-- Create "queue_locks" table
CREATE TABLE "queue_locks" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "network_id" character varying(100) NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_network_queue" to table: "queue_locks"
CREATE UNIQUE INDEX "idx_network_queue" ON "queue_locks" ("network_id");
-- Create index "idx_queue_locks_deleted_at" to table: "queue_locks"
CREATE INDEX "idx_queue_locks_deleted_at" ON "queue_locks" ("deleted_at");
-- Create "tx_queue_slots" table
CREATE TABLE "tx_queue_slots" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "network_id" character varying(100) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_janitor_cleanup" to table: "tx_queue_slots"
CREATE INDEX "idx_janitor_cleanup" ON "tx_queue_slots" ("expires_at");
-- Create index "idx_slots_count" to table: "tx_queue_slots"
CREATE INDEX "idx_slots_count" ON "tx_queue_slots" ("network_id", "expires_at");
-- Insert default network in "queue_locks" table
INSERT INTO "queue_locks" (id, network_id, created_at)
VALUES ('8300c11d-8ae2-4846-a0d8-ac34393a6ade', 'default', CURRENT_TIMESTAMP);
