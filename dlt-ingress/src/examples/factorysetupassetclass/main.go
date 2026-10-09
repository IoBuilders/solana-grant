package main

import (
	"context"
	"fmt"
	"time"

	"dlt-ingress/src/examples"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/port/crosscommand/createkey"
	"dlt-ingress/src/main/dltingress/port/crosscommand/signandsend"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

// Runs the full asset class onboarding flow against the Factory program in a
// single execution: create_asset_class, init_asset_class_version,
// enable_asset_class_version_functionalities (with every functionality
// active) and finalize_asset_class_version.
//
// Requires the Factory account to already be initialized
// (factoryinitialize) and FACTORY_MANAGER_DLT_ACCOUNT_ID to be set to the
// manager it registered. Unlike the single-instruction factory examples,
// this one keeps config_id/version/owner in memory and reuses them across
// all four calls in the same run, since they can't be shared via env vars
// between separate process executions.
func main() {
	ctx := context.Background()
	dltIngressModule := examples.InitModule(ctx)

	// 0. Reuse the manager registered by factoryinitialize, and create a new
	// FACTORY_MANAGER_DLT_ACCOUNT_ID
	managerDltAccountId := "D2qrES3S94xbSKYyBg6jwUtBDufuzwz7HeCa9h4Q3UTa"
	cmd := &createkeycross.CrossCommand{Dlt: "SVM"}
	response, err := dltIngressModule.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	ownerDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId
	ownerPubKey := solana.MustPublicKeyFromBase58(ownerDltAccountId)

	// 1. Request SOL to manager and owner accounts
	rpcClient := rpc.New(config.DltIngressConfig.DltIngress.Networks[0].Url)
	for _, dltAccountId := range []string{managerDltAccountId, ownerDltAccountId} {
		_, err = rpcClient.RequestAirdrop(
			ctx,
			solana.MustPublicKeyFromBase58(dltAccountId),
			1000*solana.LAMPORTS_PER_SOL,
			rpc.CommitmentFinalized,
		)
		if err != nil {
			panic(fmt.Errorf("failed to airdrop account: %w", err))
		}
	}

	factoryProgramId := "FEY9E77nH7R1gLGNxkhYKchJpB6MgpMrWMhkNXrNhzR5"
	eventAuthority, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("__event_authority")},
		solana.MustPublicKeyFromBase58(factoryProgramId),
	)
	if err != nil {
		panic(fmt.Errorf("failed to calculate event_authority account: %w", err))
	}

	configId := uint64(2)
	version := uint64(1)

	// Every functionality defined in programs/common/src/functionalities.rs,
	// enabled at once. The borsh vec<u16> encoder expects []any of uint64.
	allFunctionalities := make([]any, 27)
	for i := range allFunctionalities {
		allFunctionalities[i] = uint64(i)
	}

	// 2. create_asset_class
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
		SenderDltAccountId:   managerDltAccountId,
		SignersDltAccountIds: []string{managerDltAccountId},
		SmartContractId:      factoryProgramId,
		SmartContractName:    "Factory",
		MethodName:           "create_asset_class",
		MethodArgs: map[string]any{
			"manager":         managerDltAccountId,
			"config_id":       configId,
			"owner":           ownerPubKey,
			"event_authority": eventAuthority.String(),
			"program":         factoryProgramId,
		},
		NetworkId: "solana-localnet",
	})
	if err != nil {
		panic(fmt.Errorf("failed to create asset class: %w", err))
	}
	signature := response.(*signandsendcross.CrossResponse).TxId
	if err := examples.WaitForConfirmation(ctx, rpcClient, signature, 30*time.Second); err != nil {
		panic(fmt.Errorf("create asset class transaction did not confirm: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf("Asset class created, signature: %s", signature))

	// 3. init_asset_class_version
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
		SenderDltAccountId:   ownerDltAccountId,
		SignersDltAccountIds: []string{ownerDltAccountId},
		SmartContractId:      factoryProgramId,
		SmartContractName:    "Factory",
		MethodName:           "init_asset_class_version",
		MethodArgs: map[string]any{
			"owner":           ownerDltAccountId,
			"config_id":       configId,
			"version":         version,
			"event_authority": eventAuthority.String(),
			"program":         factoryProgramId,
		},
		NetworkId: "solana-localnet",
	})
	if err != nil {
		panic(fmt.Errorf("failed to init asset class version: %w", err))
	}
	signature = response.(*signandsendcross.CrossResponse).TxId
	if err := examples.WaitForConfirmation(ctx, rpcClient, signature, 30*time.Second); err != nil {
		panic(fmt.Errorf("init asset class version transaction did not confirm: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf("Asset class version initialized, signature: %s", signature))

	// 4. enable_asset_class_version_functionalities
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
		SenderDltAccountId:   ownerDltAccountId,
		SignersDltAccountIds: []string{ownerDltAccountId},
		SmartContractId:      factoryProgramId,
		SmartContractName:    "Factory",
		MethodName:           "enable_asset_class_version_functionalities",
		MethodArgs: map[string]any{
			"owner":           ownerDltAccountId,
			"config_id":       configId,
			"version":         version,
			"functionalities": allFunctionalities,
			"event_authority": eventAuthority.String(),
			"program":         factoryProgramId,
		},
		NetworkId: "solana-localnet",
	})
	if err != nil {
		panic(fmt.Errorf("failed to enable asset class version functionalities: %w", err))
	}
	signature = response.(*signandsendcross.CrossResponse).TxId
	if err := examples.WaitForConfirmation(ctx, rpcClient, signature, 30*time.Second); err != nil {
		panic(fmt.Errorf("enable asset class version functionalities transaction did not confirm: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf("Asset class version functionalities enabled, signature: %s", signature))

	// 5. finalize_asset_class_version
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
		SenderDltAccountId:   ownerDltAccountId,
		SignersDltAccountIds: []string{ownerDltAccountId},
		SmartContractId:      factoryProgramId,
		SmartContractName:    "Factory",
		MethodName:           "finalize_asset_class_version",
		MethodArgs: map[string]any{
			"owner":           ownerDltAccountId,
			"config_id":       configId,
			"version":         version,
			"event_authority": eventAuthority.String(),
			"program":         factoryProgramId,
		},
		NetworkId: "solana-localnet",
	})
	if err != nil {
		panic(fmt.Errorf("failed to finalize asset class version: %w", err))
	}
	signature = response.(*signandsendcross.CrossResponse).TxId
	if err := examples.WaitForConfirmation(ctx, rpcClient, signature, 30*time.Second); err != nil {
		panic(fmt.Errorf("finalize asset class version transaction did not confirm: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf(
		"Asset class version finalized, signature: %s. Reuse this owner in the other examples:\n  export FACTORY_ASSET_CLASS_OWNER_DLT_ACCOUNT_ID=%s",
		signature, ownerDltAccountId,
	))
}
