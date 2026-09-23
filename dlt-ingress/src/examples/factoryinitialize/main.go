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

// Initializes the Factory program's singleton Factory account, appointing the
// manager that will later be able to nominate asset class managers.
func main() {
	ctx := context.Background()
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

	// 1. Create payer and manager accounts
	cmd := &createkeycross.CrossCommand{Dlt: string(common.SVM)}
	response, err := core.App.DltIngress.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	payerDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId
	response, err = core.App.DltIngress.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	managerDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId

	// 2. Request SOL to payer account
	pubKey := solana.MustPublicKeyFromBase58(payerDltAccountId)
	_, err = rpc.New(config.AppConfig.DltIngress.Networks[0].Url).RequestAirdrop(
		ctx,
		pubKey,
		1000*solana.LAMPORTS_PER_SOL,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(fmt.Errorf("failed to airdrop account: %w", err))
	}

	factoryProgramId := "FEY9E77nH7R1gLGNxkhYKchJpB6MgpMrWMhkNXrNhzR5"
	seeds := [][]byte{
		[]byte("__event_authority"),
	}
	eventAuthority, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(factoryProgramId))
	if err != nil {
		panic(fmt.Errorf("failed to calculate event_authotity account: %w", err))
	}

	// 3. Sign and send transaction
	_, err = core.App.DltIngress.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
		SenderDltAccountId:   payerDltAccountId,
		SignersDltAccountIds: []string{payerDltAccountId, managerDltAccountId},
		SmartContractId:      factoryProgramId,
		SmartContractName:    "Factory",
		MethodName:           "initialize",
		MethodArgs: map[string]any{
			"payer":           payerDltAccountId,
			"manager":         managerDltAccountId,
			"event_authority": eventAuthority.String(),
			"program":         factoryProgramId,
		},
		NetworkId: "solana-localnet",
	})
	if err != nil {
		panic(fmt.Errorf("failed to initialize factory: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf(
		"Factory initialized. Reuse this manager in the other factory examples:\n  export FACTORY_MANAGER_DLT_ACCOUNT_ID=%s",
		managerDltAccountId,
	))
}
