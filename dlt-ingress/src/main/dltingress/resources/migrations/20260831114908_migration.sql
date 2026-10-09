-- Create "faucet_wallets" table
CREATE TABLE "faucet_wallets" ("id" uuid NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "deleted_at" timestamptz NULL, "network_id" character varying(255) NOT NULL, "custody_key_id" uuid NOT NULL, "funding_amount" character varying(100) NOT NULL, "balance_threshold" character varying(100) NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "fk_faucet_wallets_custody_key" FOREIGN KEY ("custody_key_id") REFERENCES "custody_keys" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "idx_faucet_wallets_custody_key_id" to table: "faucet_wallets"
CREATE UNIQUE INDEX "idx_faucet_wallets_custody_key_id" ON "faucet_wallets" ("custody_key_id");
-- Create index "idx_faucet_wallets_deleted_at" to table: "faucet_wallets"
CREATE INDEX "idx_faucet_wallets_deleted_at" ON "faucet_wallets" ("deleted_at");
-- Create index "idx_faucet_wallets_network_id" to table: "faucet_wallets"
CREATE UNIQUE INDEX "idx_faucet_wallets_network_id" ON "faucet_wallets" ("network_id");
