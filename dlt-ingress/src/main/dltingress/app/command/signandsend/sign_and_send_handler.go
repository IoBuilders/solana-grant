package signandsend

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/app/service/txservice"
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/domain/custodykey/service"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	transactionDomain "dlt-ingress/src/main/dltingress/domain/transaction"
	"dlt-ingress/src/main/dltingress/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/port/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/port/custody"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	"dlt-ingress/src/main/dltingress/port/portcommon"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/svmtransactionrepo"
	"dlt-ingress/src/main/dltingress/port/transanctionsender"

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
	evmTransactionRepository       evmtransactionrepo.Repository
	svmTransactionRepository       svmtransactionrepo.Repository
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
	evmTransactionRepository evmtransactionrepo.Repository,
	svmTransactionRepository svmtransactionrepo.Repository,
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
	network, err := config.AppConfig.GetDltIngressNetwork(cmd.NetworkId)
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

	transaction, err := h.txService.PrepareTransaction(ctx, txservice.TransactionRequest{
		SenderDltAccountId:   cmd.SenderDltAccountId,
		SignersDltAccountIds: cmd.SignersDltAccountIds,
		SmartContractId:      cmd.SmartContractId,
		SmartContractName:    cmd.SmartContractName,
		MethodName:           cmd.MethodName,
		MethodArgs:           cmd.MethodArgs,
		NetworkId:            cmd.NetworkId,
	}, *network)
	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	signResponse, err := h.custodyProvider.Sign(ctx, &custody.SignRequest{
		Dlt:         network.Dlt,
		CustodyKeys: custodyKeys,
		Transaction: transaction,
	})
	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	var transactionSentEvent *transactionDomain.TransactionSentEvent
	switch dlt {
	case common.EVM:
		transactionSentEvent, err = h.executeEvmTransaction(ctx, cmd, network, signResponse, transactionSender, transaction)
	case common.SVM:
		transactionSentEvent, err = h.executeSvmTransaction(ctx, network, signResponse, transactionSender, transaction)
	default:
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, domainerrors.NewInvalidDltDomainError(network.Dlt)
	}

	if err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	if err := h.eventBus.Publish(ctx, transactionSentEvent); err != nil {
		h.takeFromQueue(ctx, cmd.NetworkId)
		return nil, err
	}

	return &Response{*transactionSentEvent}, nil
}

func (h *CommandHandler) executeEvmTransaction(
	ctx context.Context,
	cmd *Command,
	network *config.NetworkConfig,
	signResponse *custody.SignResponse,
	transactionSender transanctionsender.Port,
	transaction *portcommon.TransactionResponse,
) (*transactionDomain.TransactionSentEvent, error) {
	sendTransaction, err := transactionSender.SendTransaction(ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: signResponse.SignedTransaction,
		Dlt:               network.Dlt,
		NetworkId:         network.Id,
	})
	if err != nil {
		return nil, err
	}

	evmTx, err := evmtransaction.NewEvmTransaction(
		sendTransaction.TxId,
		network.Id,
		network.Url,
		network.Dlt,
		cmd.SenderDltAccountId,
		transaction.To,
		transaction.Nonce,
		transaction.Value,
		transaction.TransactionType,
		transaction.GasLimit,
		transaction.GasPrice,
		transaction.MaxPriorityFeePerGas,
		transaction.MaxFeePerGas,
		transaction.Data,
	)
	if err != nil {
		return nil, err
	}

	// The following line is using a context.Background() because we want to persist the transaction directly to the database
	// without waiting for the completion of the command handler execution. This is because we want to reduce the risk of the DLT
	// event being processed before the transaction is committed to the database.
	if err := h.evmTransactionRepository.Save(context.Background(), evmTx); err != nil {
		return nil, err
	}

	if err := h.nonceProvider.SetNonce(ctx, &nonceprovider.SetNonceRequest{
		NetworkId:    cmd.NetworkId,
		DltAccountId: cmd.SenderDltAccountId,
		Value:        transaction.Nonce,
	}); err != nil {
		return nil, err
	}

	var maxFeePerGas amount.Amount
	if transaction.MaxFeePerGas != nil {
		maxFeePerGas = *transaction.MaxFeePerGas
	}
	evmEvtModel := transactionDomain.NewEvmTransactionEventModel(
		cmd.SenderDltAccountId,
		transaction.To,
		*transaction.Nonce,
		transaction.Value,
		transaction.Data,
		network.TransactionType,
		network.GasLimit,
		network.GasPrice,
		network.MaxPriorityFeePerGas,
		maxFeePerGas,
	)
	evt := transactionDomain.NewEvmTransactionSentEvent(evmTx.TxId, evmEvtModel)
	return evt, nil
}

func (h *CommandHandler) executeSvmTransaction(
	ctx context.Context,
	network *config.NetworkConfig,
	signResponse *custody.SignResponse,
	transactionSender transanctionsender.Port,
	transaction *portcommon.TransactionResponse,
) (*transactionDomain.TransactionSentEvent, error) {
	sendTransaction, err := transactionSender.SendTransaction(ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: signResponse.SignedTransaction,
		Dlt:               network.Dlt,
		NetworkId:         network.Id,
	})
	if err != nil {
		return nil, err
	}

	svmTx, err := svmtransaction.NewSvmTransaction(
		sendTransaction.TxId,
		network.Id,
		network.Url,
		transaction.FeePayer,
		transaction.RecentBlockhash,
		transaction.SerializedTransaction,
		transaction.CuLimit,
		transaction.CuPrice,
	)
	if err != nil {
		return nil, err
	}

	// The following line is using a context.Background() because we want to persist the transaction directly to the database
	// without waiting for the completion of the command handler execution. This is because we want to reduce the risk of the DLT
	// event being processed before the transaction is committed to the database.
	if err := h.svmTransactionRepository.Save(context.Background(), svmTx); err != nil {
		return nil, err
	}

	svmEvtModel := transactionDomain.NewSvmTransactionEventModel(
		transaction.FeePayer,
		transaction.RecentBlockhash,
		transaction.SerializedTransaction,
		transaction.CuLimit,
		transaction.CuPrice,
	)
	evt := transactionDomain.NewSvmTransactionSentEvent(svmTx.TxId, svmEvtModel)

	return evt, nil
}

func (h *CommandHandler) takeFromQueue(ctx context.Context, networkId string) {
	if err := h.boundedBlockingQueue.Take(ctx, networkId); err != nil {
		logger.ErrorWithCtx(ctx, "error taking an element from the queue when sending transaction", "error", err)
	}
}
