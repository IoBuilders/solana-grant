-- Create "solana_contract_events" table
CREATE TABLE "public"."solana_contract_events" (
  "id" uuid NOT NULL,
  "node_id" uuid NOT NULL,
  "event_name" character varying(255) NOT NULL,
  "status" character varying(20) NOT NULL,
  "program_id" character varying(44) NOT NULL,
  "signature" character varying(2000) NOT NULL,
  "slot" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "contract_event_parameters" table
CREATE TABLE "public"."contract_event_parameters" (
  "id" uuid NOT NULL,
  "type" character varying(20) NOT NULL,
  "position" bigint NOT NULL,
  "value" jsonb NOT NULL,
  "contract_event_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_solana_contract_events_parameters" FOREIGN KEY ("contract_event_id") REFERENCES "public"."solana_contract_events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
