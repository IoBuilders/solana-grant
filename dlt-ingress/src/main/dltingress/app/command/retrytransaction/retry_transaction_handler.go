package retrytransaction

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/domain/custodykey"
	"dlt-ingress/src/main/dltingress/domain/custodykey/service"
	transactionDomain "dlt-ingress/src/main/dltingress/domain/transaction"
	"dlt-ingress/src/main/dltingress/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/port/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/port/custody"
	"dlt-ingress/src/main/dltingress/port/evm"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	"dlt-ingress/src/main/dltingress/port/portcommon"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/port/transanctionsender"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type CommandHandler struct {
	custodyProvider           custody.Port
	nonceProvider             nonceprovider.Port
	transactionSenderRegistry transanctionsender.Registry
	evmTransactionRepository  evmtransactionrepo.Repository
	eventBus                  event.Bus
	evmClientRegistry         evm.ClientRegistry
	boundedBlockingQueue      boundedblockingqueue.Port
	custodyKeyExistsService   servicecustodykey.ExistsInterface
}

func NewCommandHandler(
	eventBus event.Bus,
	custodyProvider custody.Port,
	nonceProvider nonceprovider.Port,
	transactionSenderRegistry transanctionsender.Registry,
	evmTransactionRepository evmtransactionrepo.Repository,
	evmClientRegistry evm.ClientRegistry,
	boundedBlockingQueue boundedblockingqueue.Port,
	custodyKeyExistsService servicecustodykey.ExistsInterface,
) *CommandHandler {
	return &CommandHandler{
		eventBus:                  eventBus,
		custodyProvider:           custodyProvider,
		nonceProvider:             nonceProvider,
		transactionSenderRegistry: transactionSenderRegistry,
		evmTransactionRepository:  evmTransactionRepository,
		evmClientRegistry:         evmClientRegistry,
		boundedBlockingQueue:      boundedBlockingQueue,
		custodyKeyExistsService:   custodyKeyExistsService,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	oldTx, err := h.evmTransactionRepository.FindByTxId(ctx, cmd.TransactionHash)
	if err != nil {
		return nil, err
	}

	network, err := config.AppConfig.GetDltIngressNetwork(oldTx.NetworkId)
	if err != nil {
		return nil, err
	}

	err = h.boundedBlockingQueue.Put(ctx, network.Id, uuid.New())
	if err != nil {
		return nil, err
	}

	custodyKey, err := h.custodyKeyExistsService.Execute(ctx, oldTx.FromAddress)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	retryTx, errDomain := h.buildRetryTransaction(ctx, cmd, oldTx, network)
	if errDomain != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, errDomain
	}

	signResp, err := h.signTransaction(ctx, retryTx, network, custodyKey)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	transactionSender, err := h.transactionSenderRegistry.GetTransactionSender(retryTx.Dlt)
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	sendResp, err := transactionSender.SendTransaction(ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: signResp.SignedTransaction,
		Dlt:               network.Dlt,
		NetworkId:         network.Id,
	})
	if err != nil {
		h.takeFromQueue(ctx, network.Id)
		return nil, err
	}

	// The following line is using a context.Background() because we want to persist the transaction directly to the database
	// without waiting for the completion of the command handler execution. This is because we want to reduce the risk of the DLT
	// event being processed before the transaction is committed to the database.
	if err := h.saveRetriedTransaction(context.Background(), sendResp.TxId, retryTx); err != nil {
		return nil, err
	}

	evt := transactionDomain.RetriedEvent{
		BaseEvent: *event.NewBaseEvent(),
		OldTxId:   cmd.TransactionHash,
		NewTxId:   sendResp.TxId,
	}
	if err := h.eventBus.Publish(ctx, evt); err != nil {
		return nil, err
	}

	return &Response{evt}, nil
}

func (h *CommandHandler) signTransaction(ctx context.Context, tx *evmtransaction.EvmTransaction, network *config.NetworkConfig, custodyKey *custodykey.CustodyKey) (*custody.SignResponse, error) {
	transactionResponse := &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			TransactionType:      tx.TransactionType,
			ChainId:              &network.ChainId,
			Nonce:                tx.Nonce,
			GasLimit:             tx.GasLimit,
			GasPrice:             tx.GasPrice,
			MaxPriorityFeePerGas: tx.MaxPriorityFeePerGas,
			MaxFeePerGas:         tx.MaxFeePerGas,
			Data:                 tx.Data,
			Value:                tx.Value,
			To:                   tx.ToAddress,
		},
	}

	sign, err := h.custodyProvider.Sign(ctx, &custody.SignRequest{
		Dlt:         string(tx.Dlt),
		CustodyKeys: []*custodykey.CustodyKey{custodyKey},
		Transaction: transactionResponse,
	})
	if err != nil {
		return nil, err
	}
	return sign, nil
}

func (h *CommandHandler) buildRetryTransaction(
	ctx context.Context,
	cmd *Command,
	oldTx *evmtransaction.EvmTransaction,
	network *config.NetworkConfig,
) (*evmtransaction.EvmTransaction, error) {
	newTx := oldTx.Clone()
	if cmd.GasLimit != nil {
		newTx.GasLimit = cmd.GasLimit
	}

	newNonce, err := h.findNonce(ctx, cmd, network.Id, oldTx.FromAddress)
	if err != nil {
		return nil, err
	}
	newTx.Nonce = newNonce

	if oldTx.OriginalTxId != nil {
		newTx.OriginalTxId = oldTx.OriginalTxId
	} else {
		newTx.OriginalTxId = &oldTx.TxId
	}

	switch oldTx.TransactionType {
	case portcommon.TransactionTypeLegacy:
		if cmd.GasPrice != nil {
			newTx.GasPrice = cmd.GasPrice
		} else if cmd.UseMultiplier {
			multiplied := newTx.GasPrice.MulWithDecimals(network.FeeMultiplier, 0)
			newTx.GasPrice = &multiplied
		}

	case portcommon.TransactionTypeDynamicFee:
		if cmd.MaxPriorityFeePerGas != nil {
			newTx.MaxPriorityFeePerGas = cmd.MaxPriorityFeePerGas
		} else if cmd.UseMultiplier {
			multiplied := newTx.MaxPriorityFeePerGas.MulWithDecimals(network.FeeMultiplier, 0)
			newTx.MaxPriorityFeePerGas = &multiplied
		}

		evmClient, err := h.evmClientRegistry.GetClientForNetworkId(ctx, network.Id)
		if err != nil {
			return nil, err
		}

		baseFeePerGas, err := evmClient.GetBaseFeePerGas(ctx)
		if err != nil {
			return nil, err
		}

		incrementedBaseFee := baseFeePerGas.MulWithDecimals(network.FeeMultiplier, 0)
		maxFeePerGas := incrementedBaseFee.Add(*newTx.MaxPriorityFeePerGas)
		newTx.MaxFeePerGas = &maxFeePerGas
	}

	return newTx, nil
}

func (h *CommandHandler) findNonce(
	ctx context.Context,
	cmd *Command,
	networkId string,
	dltAccountId string,
) (*amount.Amount, error) {
	if cmd.Nonce != nil {
		return cmd.Nonce, nil
	} else {
		remoteNonce, err := h.nonceProvider.GetNonce(ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    networkId,
			DltAccountId: dltAccountId,
		})
		if err != nil {
			return nil, err
		}
		return remoteNonce, nil
	}
}

func (h *CommandHandler) saveRetriedTransaction(ctx context.Context, txId string, retryTx *evmtransaction.EvmTransaction) error {
	persistedTx, errDomain := evmtransaction.NewEvmTransaction(
		txId,
		retryTx.NetworkId,
		retryTx.NetworkUrl,
		string(retryTx.Dlt),
		retryTx.FromAddress,
		retryTx.ToAddress,
		retryTx.Nonce,
		retryTx.Value,
		retryTx.TransactionType,
		retryTx.GasLimit,
		retryTx.GasPrice,
		retryTx.MaxPriorityFeePerGas,
		retryTx.MaxFeePerGas,
		retryTx.Data,
	)
	if errDomain != nil {
		return errDomain
	}
	persistedTx.OriginalTxId = retryTx.OriginalTxId

	if err := h.evmTransactionRepository.Save(ctx, persistedTx); err != nil {
		return err
	}
	return nil
}

func (h *CommandHandler) takeFromQueue(ctx context.Context, networkId string) {
	if err := h.boundedBlockingQueue.Take(ctx, networkId); err != nil {
		logger.ErrorWithCtx(ctx, "error taking an element from the queue when sending transaction", "error", err)
	}
}
