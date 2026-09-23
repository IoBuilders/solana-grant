-- Rename a constraint from "ethereum_transactions_pkey" to "evm_transactions_pkey"
ALTER TABLE "evm_transactions" RENAME CONSTRAINT "ethereum_transactions_pkey" TO "evm_transactions_pkey";
-- Modify "svm_transactions" table
ALTER TABLE "svm_transactions" ADD COLUMN "cu_limit" character varying(100) NOT NULL;
ALTER TABLE "svm_transactions" ADD COLUMN "cu_price" character varying(100) NOT NULL;
