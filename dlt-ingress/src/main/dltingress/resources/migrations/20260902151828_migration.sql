-- Modify "faucet_wallets" table
ALTER TABLE "faucet_wallets" ADD COLUMN "enabled" boolean NOT NULL DEFAULT true;
