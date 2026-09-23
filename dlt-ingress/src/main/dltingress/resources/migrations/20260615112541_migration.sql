-- Modify "event_consumers" table
ALTER TABLE "event_consumers" ADD COLUMN "error_details" character varying(5000) NULL;
