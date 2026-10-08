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

// Initializes the Factory program's singleton Factory account, appointing the
// manager that will later be able to nominate asset class managers.
func main() {
	ctx := context.Background()
	dltIngressModule := examples.InitModule(ctx)

	// 1. Create payer and manager accounts
	cmd := &createkeycross.CrossCommand{Dlt: "SVM"}
	response, err := dltIngressModule.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	payerDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, cmd)
	if err != nil {
		panic(fmt.Errorf("failed to create account: %w", err))
	}
	managerDltAccountId := response.(*createkeycross.CrossResponse).DltAccountId

	// 2. Request SOL to payer account
	rpcClient := rpc.New(config.DltIngressConfig.DltIngress.Networks[0].Url)
	pubKey := solana.MustPublicKeyFromBase58(payerDltAccountId)
	_, err = rpcClient.RequestAirdrop(
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
	response, err = dltIngressModule.CrossCommandBus.Dispatch(ctx, &signandsendcross.CrossCommand{
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
	signature := response.(*signandsendcross.CrossResponse).TxId

	if err := examples.WaitForConfirmation(ctx, rpcClient, signature, 30*time.Second); err != nil {
		panic(fmt.Errorf("factory initialize transaction did not confirm: %w", err))
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf(
		"Factory initialized, signature: %s. Reuse this manager in the other factory examples:\n  export FACTORY_MANAGER_DLT_ACCOUNT_ID=%s",
		signature, managerDltAccountId,
	))
}
