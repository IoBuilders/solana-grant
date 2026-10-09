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

func main() {
	ctx := context.Background()
	dltIngressModule := examples.InitModule(ctx)

	// 1. Create sender and mint accounts
	cmd := &createkeycross.CrossCommand{Dlt: "SVM"}
	response, err := dltIngressModule.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	senderDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	mintDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId

	// 2. Request SOL to sender account
	rpcClient := rpc.New(config.DltIngressConfig.DltIngress.Networks[0].Url)
	pubKey := solana.MustPublicKeyFromBase58(senderDltAccountId)
	_, err = rpcClient.RequestAirdrop(
		ctx,
		pubKey,
		1000*solana.LAMPORTS_PER_SOL,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(fmt.Errorf("failed to airdrop account: %w", err))
	}

	// 3. Sign and send transaction
	deployProgramId := "HCe5Um7ThFBzDSyn256EPQvyr6jy6E66ydzZ5hMta3Tq"
	seeds := [][]byte{
		[]byte("__event_authority"),
	}
	eventAuthority, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(deployProgramId))
	if err != nil {
		panic(fmt.Errorf("failed to calculate event_authotity account: %w", err))
	}
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
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
	signature := response.(*signandsendcross.CrossResponse).TxId

	if err := examples.WaitForConfirmation(ctx, rpcClient, signature, 30*time.Second); err != nil {
		panic(fmt.Errorf("deploy asset transaction did not confirm: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf("Asset deployment finished: %s, signature: %s", mintDltAccountId, signature))
}
