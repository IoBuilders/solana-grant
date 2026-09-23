package main

import (
	"context"
	"dlt-ingress"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/core"
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/port/command/createkey"
	"dlt-ingress/src/main/dltingress/port/command/signandsend"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability"
)

func main() {
	ctx := context.Background()
	// 1. Start dlt ingress application
	if err := config.LoadConfig(ctx); err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}
	_, err := observability.SetupOTelSDK(
		ctx,
		observability.WithLogLevel(config.AppConfig.LoggingConfig.Level),
		observability.WithLogFormat(config.AppConfig.LoggingConfig.Format),
		observability.WithScope(observability.NewScope("gitlab.com/iobuilders/projects/eng/o2d/dlt-ingress", dlt_ingress.Version)),
	)
	if err != nil {
		panic(fmt.Errorf("failed to setup otel sdk: %w", err))
	}
	core.StartApplication(ctx)

	// 2. Create sender and mint accounts
	cmd := &createkeycross.CrossCommand{Dlt: string(common.SVM)}
	response, err := core.App.DltIngress.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	senderDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId
	response, err = core.App.DltIngress.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	mintDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId

	// 3. Request SOL to sender account
	pubKey := solana.MustPublicKeyFromBase58(senderDltAccountId)
	_, err = rpc.New(config.AppConfig.DltIngress.Networks[0].Url).RequestAirdrop(
		ctx,
		pubKey,
		1000*solana.LAMPORTS_PER_SOL,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(fmt.Errorf("failed to airdrop account: %w", err))
	}

	// 4. Sign and send transaction
	deployProgramId := "HCe5Um7ThFBzDSyn256EPQvyr6jy6E66ydzZ5hMta3Tq"
	seeds := [][]byte{
		[]byte("__event_authority"),
	}
	eventAuthority, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(deployProgramId))
	if err != nil {
		panic(fmt.Errorf("failed to calculate event_authotity account: %w", err))
	}
	_, err = core.App.DltIngress.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
		SenderDltAccountId:   senderDltAccountId,
		SignersDltAccountIds: []string{senderDltAccountId, mintDltAccountId},
		SmartContractId:      deployProgramId,
		SmartContractName:    "Deploy",
		MethodName:           "deploy_mint",
		MethodArgs: map[string]any{
			"deployer":               senderDltAccountId,
			"payer":                  senderDltAccountId,
			"mint":                   mintDltAccountId,
			"event_authority":        eventAuthority.String(),
			"program":                deployProgramId,
			"access_control_program": "GpyjQqBWux3JYqxKCXFrDbWZmhFWBJWVaVivkBW2DL2w",
			"params": map[string]any{
				"decimals":               uint64(8),
				"name":                   "ExampleName",
				"symbol":                 "ExampleSymbol",
				"uri":                    "ExampleUri",
				"additional_metadata":    []any{},
				"asset_class_config_id":  uint64(1),
				"asset_class_version_id": uint64(1),
			},
		},
		NetworkId: "solana-localnet",
	})
	if err != nil {
		panic(fmt.Errorf("failed to deploy asset: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf(
		"Asset deployment finished. Reuse these in the other examples:\n"+
			"  export DEPLOY_MINT_DLT_ACCOUNT_ID=%s\n"+
			"  export DEPLOY_DEPLOYER_DLT_ACCOUNT_ID=%s",
		mintDltAccountId,
		senderDltAccountId,
	))
}
