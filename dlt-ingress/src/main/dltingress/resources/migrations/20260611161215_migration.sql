ALTER TABLE "ethereum_transactions" RENAME TO "evm_transactions";
ALTER INDEX "ethereum_transactions_tx_id_key" RENAME TO "evm_transactions_tx_id_key";
ALTER INDEX "idx_ethereum_transactions_deleted_at" RENAME TO "idx_evm_transactions_deleted_at";
