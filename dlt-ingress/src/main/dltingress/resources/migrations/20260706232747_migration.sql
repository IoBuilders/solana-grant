-- Modify "evm_transactions" table
ALTER TABLE "evm_transactions" ADD COLUMN "original_tx_id" character varying(255) NULL;
-- Modify "svm_transactions" table
ALTER TABLE "svm_transactions" ADD COLUMN "original_tx_id" character varying(255) NULL;
