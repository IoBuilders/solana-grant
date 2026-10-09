package signandsend

import (
	"context"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/app/service/txservice"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey/service"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/internal/infra/transanctionsender"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type CommandHandler struct {
	boundedBlockingQueue           boundedblockingqueue.Port
	nonceProvider                  nonceprovider.Port
	custodyProvider                custody.Port
	transactionSenderRegistry      transanctionsender.Registry
	custodyKeyExistMultipleService servicecustodykey.ExistMultipleInterface
	evmTransactionRepository       evmtransaction.Repository
	svmTransactionRepository       svmtransaction.Repository
	eventBus                       event.Bus
	txService                      *txservice.AppService
}

func NewCommandHandler(
	eventBus event.Bus,
	txService *txservice.AppService,
	nonceProvider nonceprovider.Port,
	custodyProvider custody.Port,
	transactionSenderRegistry transanctionsender.Registry,
	custodyKeyExistMultipleService servicecustodykey.ExistMultipleInterface,
	boundedBlockingQueue boundedblockingqueue.Port,
	evmTransactionRepository evmtransaction.Repository,
	svmTransactionRepository svmtransaction.Repository,
) *CommandHandler {
	return &CommandHandler{
		eventBus:                       eventBus,
		txService:                      txService,
		nonceProvider:                  nonceProvider,
		boundedBlockingQueue:           boundedBlockingQueue,
		custodyProvider:                custodyProvider,
		transactionSenderRegistry:      transactionSenderRegistry,
		custodyKeyExistMultipleService: custodyKeyExistMultipleService,
		evmTransactionRepository:       evmTransactionRepository,
		svmTransactionRepository:       svmTransactionRepository,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	network, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(cmd.NetworkId)
	if err != nil {
		return nil, err
	}

	err = h.boundedBlockingQueue.Put(ctx, cmd.NetworkId, uuid.New())
	if err != nil {
		return nil, err
	}

	dlt, err := common.ParseDlt(network.Dlt)
	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	custodyKeys, err := h.custodyKeyExistMultipleService.Execute(ctx, cmd.SignersDltAccountIds)
	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	transactionSender, err := h.transactionSenderRegistry.GetTransactionSender(dlt)
	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	tx, err := h.txService.PrepareTransaction(ctx, txservice.TransactionRequest{
		SenderDltAccountId:   cmd.SenderDltAccountId,
		SignersDltAccountIds: cmd.SignersDltAccountIds,
		SmartContractId:      cmd.SmartContractId,
		SmartContractName:    cmd.SmartContractName,
		MethodName:           cmd.MethodName,
		MethodArgs:           cmd.MethodArgs,
		NetworkId:            cmd.NetworkId,
		ResolveNestedCalls:   cmd.ResolveNestedCalls,
	}, *network)
	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	signResponse, err := h.custodyProvider.Sign(ctx, &custody.SignRequest{
		Dlt:         network.Dlt,
		CustodyKeys: custodyKeys,
		Transaction: tx,
	})
	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	var transactionSentEvent transaction.TransactionSentEvent
	switch dlt {
	case common.EVM:
		transactionSentEvent, err = h.executeEvmTransaction(ctx, cmd, network, signResponse, transactionSender, tx)
	case common.SVM:
		transactionSentEvent, err = h.executeSvmTransaction(ctx, network, signResponse, transactionSender, tx)
	default:
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, domainerrors.NewInvalidDltDomainError(network.Dlt)
	}

	if err != nil {
		return nil, err
	}

	if err := h.eventBus.Publish(ctx, transactionSentEvent); err != nil {
		return nil, err
	}

	return &Response{transactionSentEvent}, nil
}

func (h *CommandHandler) executeEvmTransaction(
	ctx context.Context,
	cmd *Command,
	network *config.NetworkConfig,
	signResponse *custody.SignResponse,
	transactionSender transanctionsender.Port,
	transactionResponse *portcommon.TransactionResponse,
) (transaction.TransactionSentEvent, error) {
	evmTx, err := evmtransaction.NewEvmTransaction(
		signResponse.TxId,
		network.Id,
		network.Url,
		network.Dlt,
		cmd.SenderDltAccountId,
		transactionResponse.To,
		transactionResponse.Nonce,
		transactionResponse.Value,
		transactionResponse.TransactionType,
		transactionResponse.GasLimit,
		transactionResponse.GasPrice,
		transactionResponse.MaxPriorityFeePerGas,
		transactionResponse.MaxFeePerGas,
		transactionResponse.Data,
	)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return transaction.TransactionSentEvent{}, err
	}

	// The following line is using a context.Background() because we want to persist the transactionResponse directly to the database
	// without waiting for the completion of the command handler execution. This is because we want to reduce the risk of the DLT
	// event being processed before the transactionResponse is committed to the database. The transactionResponse id is deterministic and known
	// before sending, so we persist it before sending it to the node to close the race window entirely.
	if err := h.evmTransactionRepository.Save(context.Background(), evmTx); err != nil {
		h.takeFromQueue(ctx, network.Id)
		return transaction.TransactionSentEvent{}, err
	}

	if _, err := transactionSender.SendTransaction(ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: signResponse.SignedTransaction,
		Dlt:               network.Dlt,
		NetworkId:         network.Id,
	}); err != nil {
		h.deleteEvmTransaction(context.Background(), evmTx)
		h.takeFromQueue(ctx, network.Id)
		return transaction.TransactionSentEvent{}, err
	}

	if err := h.nonceProvider.SetNonce(ctx, &nonceprovider.SetNonceRequest{
		NetworkId:    cmd.NetworkId,
		DltAccountId: cmd.SenderDltAccountId,
		Value:        transactionResponse.Nonce,
	}); err != nil {
		return transaction.TransactionSentEvent{}, err
	}

	var maxFeePerGas amount.Amount = *amount.Zero()
	if transactionResponse.MaxFeePerGas != nil {
		maxFeePerGas = *transactionResponse.MaxFeePerGas
	}
	evmEvtModel := transaction.NewEvmTransactionEventModel(
		cmd.SenderDltAccountId,
		transactionResponse.To,
		*transactionResponse.Nonce,
		transactionResponse.Value,
		transactionResponse.Data,
		network.TransactionType,
		network.GasLimit,
		network.GasPrice,
		network.MaxPriorityFeePerGas,
		maxFeePerGas,
	)
	evt := transaction.NewEvmTransactionSentEvent(evmTx.TxId, network.Id, evmEvtModel)
	return evt, nil
}

func (h *CommandHandler) executeSvmTransaction(
	ctx context.Context,
	network *config.NetworkConfig,
	signResponse *custody.SignResponse,
	transactionSender transanctionsender.Port,
	transactionResponse *portcommon.TransactionResponse,
) (transaction.TransactionSentEvent, error) {
	svmTx, err := svmtransaction.NewSvmTransaction(
		signResponse.TxId,
		network.Id,
		network.Url,
		transactionResponse.FeePayer,
		transactionResponse.RecentBlockhash,
		transactionResponse.SerializedTransaction,
		transactionResponse.CuLimit,
		transactionResponse.CuPrice,
	)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return transaction.TransactionSentEvent{}, err
	}

	// The following line is using a context.Background() because we want to persist the transactionResponse directly to the database
	// without waiting for the completion of the command handler execution. This is because we want to reduce the risk of the DLT
	// event being processed before the transactionResponse is committed to the database. The transactionResponse id is deterministic and known
	// before sending, so we persist it before sending it to the node to close the race window entirely.
	if err := h.svmTransactionRepository.Save(context.Background(), svmTx); err != nil {
		h.takeFromQueue(ctx, network.Id)
		return transaction.TransactionSentEvent{}, err
	}

	if _, err := transactionSender.SendTransaction(ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: signResponse.SignedTransaction,
		Dlt:               network.Dlt,
		NetworkId:         network.Id,
	}); err != nil {
		h.deleteSvmTransaction(context.Background(), svmTx)
		h.takeFromQueue(ctx, network.Id)
		return transaction.TransactionSentEvent{}, err
	}

	svmEvtModel := transaction.NewSvmTransactionEventModel(
		transactionResponse.FeePayer,
		transactionResponse.RecentBlockhash,
		transactionResponse.SerializedTransaction,
		transactionResponse.CuLimit,
		transactionResponse.CuPrice,
	)
	evt := transaction.NewSvmTransactionSentEvent(svmTx.TxId, network.Id, svmEvtModel)

	return evt, nil
}

func (h *CommandHandler) takeFromQueue(ctx context.Context, networkId string) {
	if err := h.boundedBlockingQueue.Take(ctx, networkId); err != nil {
		logger.ErrorWithCtx(ctx, "error taking an element from the queue when sending transaction", "error", err)
	}
}

func (h *CommandHandler) deleteEvmTransaction(ctx context.Context, tx *evmtransaction.EvmTransaction) {
	if err := h.evmTransactionRepository.HardDelete(ctx, tx); err != nil {
		logger.ErrorWithCtx(ctx, "error deleting transaction after failed send", "error", err, "txId", tx.TxId)
	}
}

func (h *CommandHandler) deleteSvmTransaction(ctx context.Context, tx *svmtransaction.SvmTransaction) {
	if err := h.svmTransactionRepository.HardDelete(ctx, tx); err != nil {
		logger.ErrorWithCtx(ctx, "error deleting transaction after failed send", "error", err, "txId", tx.TxId)
	}
}
