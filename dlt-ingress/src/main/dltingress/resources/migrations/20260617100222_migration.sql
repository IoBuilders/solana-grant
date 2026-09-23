-- Modify "failed_transactions" table
ALTER TABLE "failed_transactions" ALTER COLUMN "error_details" SET NOT NULL;
