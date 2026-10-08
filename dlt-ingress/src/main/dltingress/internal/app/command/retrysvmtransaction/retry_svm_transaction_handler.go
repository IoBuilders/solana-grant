package retrysvmtransaction

import (
	"context"
	"encoding/base64"
	"math"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey/service"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transanctionsender"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type CommandHandler struct {
	custodyProvider                custody.Port
	svmClientRegistry              svm.ClientRegistry
	transactionSenderRegistry      transanctionsender.Registry
	svmTransactionRepository       svmtransaction.Repository
	eventBus                       event.Bus
	boundedBlockingQueue           boundedblockingqueue.Port
	custodyKeyExistMultipleService servicecustodykey.ExistMultipleInterface
}

func NewCommandHandler(
	eventBus event.Bus,
	custodyProvider custody.Port,
	svmClientRegistry svm.ClientRegistry,
	transactionSenderRegistry transanctionsender.Registry,
	svmTransactionRepository svmtransaction.Repository,
	boundedBlockingQueue boundedblockingqueue.Port,
	custodyKeyExistMultipleService servicecustodykey.ExistMultipleInterface,
) *CommandHandler {
	return &CommandHandler{
		eventBus:                       eventBus,
		custodyProvider:                custodyProvider,
		svmClientRegistry:              svmClientRegistry,
		transactionSenderRegistry:      transactionSenderRegistry,
		svmTransactionRepository:       svmTransactionRepository,
		boundedBlockingQueue:           boundedBlockingQueue,
		custodyKeyExistMultipleService: custodyKeyExistMultipleService,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	oldTx, err := h.svmTransactionRepository.FindByTxId(ctx, cmd.TransactionHash)
	if err != nil {
		return nil, err
	}

	network, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(oldTx.NetworkId)
	if err != nil {
		return nil, err
	}

	if err := h.boundedBlockingQueue.Put(ctx, network.Id, uuid.New()); err != nil {
		return nil, err
	}

	oldTxDecoded, err := solana.TransactionFromBase64(oldTx.SerializedTransaction)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, domainerrors.NewInvalidStoredSvmTransactionDomainError(err)
	}

	// A transaction can require signatures beyond the fee payer's (e.g. a fresh account being
	// created, like deploy_mint's `mint`). All of them must be re-signed on retry too, or the
	// rebuilt transaction fails signature verification with the original (now stale) signatures
	// missing from the signer slots the rebuild still reserves for them.
	signerIds := make([]string, len(oldTxDecoded.Message.Signers()))
	for i, pk := range oldTxDecoded.Message.Signers() {
		signerIds[i] = pk.String()
	}

	custodyKeys, err := h.custodyKeyExistMultipleService.Execute(ctx, signerIds)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	retryTx, err := h.buildRetryTransaction(ctx, cmd, oldTx, oldTxDecoded, network)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	signResp, err := h.custodyProvider.Sign(ctx, &custody.SignRequest{
		Dlt:         string(oldTx.Dlt),
		CustodyKeys: custodyKeys,
		Transaction: &portcommon.TransactionResponse{SVMTransactionResponse: retryTx},
	})
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	transactionSender, err := h.transactionSenderRegistry.GetTransactionSender(oldTx.Dlt)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	// The following line is using a context.Background() because we want to persist the transaction directly to the database
	// without waiting for the completion of the command handler execution. This is because we want to reduce the risk of the DLT
	// event being processed before the transaction is committed to the database. The transaction id is deterministic and known
	// before sending, so we persist it before sending it to the node to close the race window entirely.
	persistedTx, err := h.saveRetriedTransaction(context.Background(), signResp.TxId, oldTx, retryTx)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	if _, err := transactionSender.SendTransaction(ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: signResp.SignedTransaction,
		Dlt:               network.Dlt,
		NetworkId:         network.Id,
	}); err != nil {
		h.deleteRetriedTransaction(context.Background(), persistedTx)
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	evt := transaction.RetriedEvent{
		BaseEvent: *event.NewBaseEvent(),
		OldTxId:   cmd.TransactionHash,
		NewTxId:   signResp.TxId,
	}
	if err := h.eventBus.Publish(ctx, evt); err != nil {
		return nil, err
	}

	return &Response{evt}, nil
}

// buildRetryTransaction decodes the stored unsigned transaction, swaps in a fresh blockhash (Solana's
// replay-protection equivalent of a nonce, so unlike EVM it is never left to the caller), and rebuilds
// the compute-budget instructions with the resolved CuLimit/CuPrice. The rest of the instructions are
// carried over unchanged: a retry resubmits the same call, it does not rebuild it from the IDL.
func (h *CommandHandler) buildRetryTransaction(
	ctx context.Context,
	cmd *Command,
	oldTx *svmtransaction.SvmTransaction,
	oldTxDecoded *solana.Transaction,
	network *config.NetworkConfig,
) (*portcommon.SVMTransactionResponse, error) {
	client, err := h.svmClientRegistry.GetClientForNetworkId(network.Id)
	if err != nil {
		return nil, err
	}
	blockhashStr, err := client.GetRecentBlockhash(ctx)
	if err != nil {
		return nil, err
	}
	blockhash, err := solana.HashFromBase58(blockhashStr)
	if err != nil {
		return nil, err
	}

	feePayer, err := solana.PublicKeyFromBase58(oldTx.FeePayer)
	if err != nil {
		return nil, err
	}

	kept, err := decompileNonComputeBudgetInstructions(oldTxDecoded)
	if err != nil {
		return nil, err
	}

	cuLimit := h.resolveCuLimit(ctx, cmd, oldTx, network)
	cuPrice := h.resolveCuPrice(cmd, oldTx, network)

	instructions := []solana.Instruction{}
	if cuLimit != nil && !cuLimit.IsZero() {
		cuLimitVal := cuLimit.RawValue().Uint64()
		if cuLimitVal > math.MaxUint32 {
			return nil, domainerrors.NewComputeUnitLimitOverflowDomainError(cuLimitVal, uint32(math.MaxUint32))
		}
		instructions = append(instructions, computebudget.NewSetComputeUnitLimitInstruction(uint32(cuLimitVal)).Build())
	}
	if cuPrice != nil && !cuPrice.IsZero() {
		instructions = append(instructions, computebudget.NewSetComputeUnitPriceInstruction(cuPrice.RawValue().Uint64()).Build())
	}
	instructions = append(instructions, kept...)

	newTx, err := solana.NewTransaction(instructions, blockhash, solana.TransactionPayer(feePayer))
	if err != nil {
		return nil, err
	}

	return buildSVMTransactionResponse(newTx, feePayer, blockhashStr, cuLimit, cuPrice)
}

// decompileNonComputeBudgetInstructions turns the compiled instructions of an already-built transaction
// back into generic solana.Instruction values, dropping any existing ComputeBudget program instructions
// so they can be rebuilt from the resolved CuLimit/CuPrice instead of carried over stale.
func decompileNonComputeBudgetInstructions(tx *solana.Transaction) ([]solana.Instruction, error) {
	kept := make([]solana.Instruction, 0, len(tx.Message.Instructions))
	for _, ci := range tx.Message.Instructions {
		programID, err := tx.ResolveProgramIDIndex(ci.ProgramIDIndex)
		if err != nil {
			return nil, err
		}
		if programID.Equals(solana.ComputeBudget) {
			continue
		}
		accounts, err := ci.ResolveInstructionAccounts(&tx.Message)
		if err != nil {
			return nil, err
		}
		kept = append(kept, solana.NewInstruction(programID, accounts, ci.Data))
	}
	return kept, nil
}

func buildSVMTransactionResponse(
	tx *solana.Transaction,
	feePayer solana.PublicKey,
	blockhash string,
	cuLimit *amount.Amount,
	cuPrice *amount.Amount,
) (*portcommon.SVMTransactionResponse, error) {
	raw, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if len(raw) > svmtransaction.MaxSvmTransactionSize {
		return nil, domainerrors.NewSvmTransactionSizeExceedsLimitDomainError(len(raw), svmtransaction.MaxSvmTransactionSize)
	}
	serialized := base64.StdEncoding.EncodeToString(raw)

	accountKeys := make([]string, len(tx.Message.AccountKeys))
	for i, ak := range tx.Message.AccountKeys {
		accountKeys[i] = ak.String()
	}

	instructions := make([]portcommon.SVMCompiledInstruction, len(tx.Message.Instructions))
	for i, ci := range tx.Message.Instructions {
		instructions[i] = portcommon.SVMCompiledInstruction{
			ProgramIDIndex: ci.ProgramIDIndex,
			AccountIndices: ci.Accounts,
			Data:           base64.StdEncoding.EncodeToString(ci.Data),
		}
	}

	return &portcommon.SVMTransactionResponse{
		SerializedTransaction: serialized,
		FeePayer:              feePayer.String(),
		RecentBlockhash:       blockhash,
		CuLimit:               cuLimit,
		CuPrice:               cuPrice,
		Header: portcommon.SVMMessageHeader{
			NumRequiredSignatures:       tx.Message.Header.NumRequiredSignatures,
			NumReadonlySignedAccounts:   tx.Message.Header.NumReadonlySignedAccounts,
			NumReadonlyUnsignedAccounts: tx.Message.Header.NumReadonlyUnsignedAccounts,
		},
		AccountKeys:  accountKeys,
		Instructions: instructions,
	}, nil
}

// resolveCuLimit mirrors the EVM retry's GasLimit handling: an explicit override is applied as given
// (this is an admin action, so the relay is what rejects it on send), with only a warning logged when
// it exceeds the configured ceiling. Omitted, the original transaction's limit is reused unchanged.
func (h *CommandHandler) resolveCuLimit(
	ctx context.Context,
	cmd *Command,
	oldTx *svmtransaction.SvmTransaction,
	network *config.NetworkConfig,
) *amount.Amount {
	if cmd.CuLimit == nil {
		return oldTx.CuLimit
	}
	if network.HasCuLimitCap() && cmd.CuLimit.GreaterThan(network.MaxCuLimit) {
		logger.WarnWithCtx(ctx, "retry compute unit limit exceeds the configured network ceiling, applying it anyway",
			"requested", cmd.CuLimit.String(), "ceiling", network.MaxCuLimit.String(), "txId", oldTx.TxId)
	}
	return cmd.CuLimit
}

// resolveCuPrice mirrors the EVM retry's GasPrice handling: an explicit override wins, otherwise
// useMultiplier bumps the original compute unit price by the configured fee multiplier, otherwise
// the original value is reused unchanged.
func (h *CommandHandler) resolveCuPrice(cmd *Command, oldTx *svmtransaction.SvmTransaction, network *config.NetworkConfig) *amount.Amount {
	if cmd.CuPrice != nil {
		return cmd.CuPrice
	}
	if cmd.UseMultiplier && oldTx.CuPrice != nil {
		multiplied := oldTx.CuPrice.MulWithDecimals(network.FeeMultiplier, 0)
		return &multiplied
	}
	return oldTx.CuPrice
}

func (h *CommandHandler) saveRetriedTransaction(
	ctx context.Context,
	txId string,
	oldTx *svmtransaction.SvmTransaction,
	retryTx *portcommon.SVMTransactionResponse,
) (*svmtransaction.SvmTransaction, error) {
	persistedTx, err := svmtransaction.NewSvmTransaction(
		txId,
		oldTx.NetworkId,
		oldTx.NetworkUrl,
		retryTx.FeePayer,
		retryTx.RecentBlockhash,
		retryTx.SerializedTransaction,
		retryTx.CuLimit,
		retryTx.CuPrice,
	)
	if err != nil {
		return nil, err
	}
	if oldTx.OriginalTxId != nil {
		persistedTx.OriginalTxId = oldTx.OriginalTxId
	} else {
		persistedTx.OriginalTxId = &oldTx.TxId
	}

	if err := h.svmTransactionRepository.Save(ctx, persistedTx); err != nil {
		return nil, err
	}
	return persistedTx, nil
}

func (h *CommandHandler) takeFromQueue(ctx context.Context, networkId string) {
	if err := h.boundedBlockingQueue.Take(ctx, networkId); err != nil {
		logger.ErrorWithCtx(ctx, "error taking an element from the queue when sending transaction", "error", err)
	}
}

func (h *CommandHandler) deleteRetriedTransaction(ctx context.Context, tx *svmtransaction.SvmTransaction) {
	if err := h.svmTransactionRepository.HardDelete(ctx, tx); err != nil {
		logger.ErrorWithCtx(ctx, "error deleting transaction after failed send", "error", err, "txId", tx.TxId)
	}
}
