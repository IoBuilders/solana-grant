package txservice

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/port/evm"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	"dlt-ingress/src/main/dltingress/port/portcommon"
	"dlt-ingress/src/main/dltingress/port/svm"
	"dlt-ingress/src/main/dltingress/port/transactiongasestimator"
	"math/big"
	"sort"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type TransactionRequest struct {
	SenderDltAccountId   string
	SignersDltAccountIds []string
	SmartContractId      string
	SmartContractName    string
	MethodName           string
	MethodArgs           map[string]any
	NetworkId            string
}

type AppService struct {
	nonceProvider                    nonceprovider.Port
	blockhashProvider                svm.BlockhashProvider
	transactionBuilderRegistry       contracttransactionbuilder.Registry
	evmClientRegistry                evm.ClientRegistry
	svmClientRegistry                svm.ClientRegistry
	transactionGasEstimationRegistry transactiongasestimator.Registry
}

func NewAppService(
	nonceProvider nonceprovider.Port,
	blockhashProvider svm.BlockhashProvider,
	transactionBuilderRegistry contracttransactionbuilder.Registry,
	evmClientRegistry evm.ClientRegistry,
	svmClientRegistry svm.ClientRegistry,
	transactionGasEstimationRegistry transactiongasestimator.Registry,
) *AppService {
	return &AppService{
		nonceProvider:                    nonceProvider,
		blockhashProvider:                blockhashProvider,
		transactionBuilderRegistry:       transactionBuilderRegistry,
		evmClientRegistry:                evmClientRegistry,
		svmClientRegistry:                svmClientRegistry,
		transactionGasEstimationRegistry: transactionGasEstimationRegistry,
	}
}

func (s *AppService) PrepareTransaction(ctx context.Context, req TransactionRequest, network config.NetworkConfig) (*portcommon.TransactionResponse, error) {
	dlt, err := common.ParseDlt(network.Dlt)
	if err != nil {
		return nil, err
	}

	transactionBuilder, err := s.transactionBuilderRegistry.GetContractTransactionBuilder(dlt, req.SmartContractName)
	if err != nil {
		return nil, err
	}

	var transactionRequest *contracttransactionbuilder.BuildTransactionRequest
	switch dlt {
	case common.EVM:
		transactionRequest, err = s.buildEvmTransactionRequest(ctx, req, network)
	case common.SVM:
		transactionRequest, err = s.buildSvmTransactionRequest(ctx, req, network)
	default:
		return nil, domainerrors.NewInvalidDltDomainError(network.Dlt)
	}

	if err != nil {
		return nil, err
	}

	transaction, err := transactionBuilder.BuildTransaction(transactionRequest)
	if err != nil {
		return nil, err
	}

	transactionWithGasOverridden, err := s.estimateGasAndOverrideLimit(ctx, dlt, req.SenderDltAccountId, network, transactionBuilder, transactionRequest, transaction)
	if err != nil {
		logger.WarnWithCtx(ctx, "error estimating gas", "error", err)
		return transaction, nil
	}

	return transactionWithGasOverridden, nil
}

func (s *AppService) estimateGasAndOverrideLimit(
	ctx context.Context,
	dlt common.Dlt,
	from string,
	network config.NetworkConfig,
	transactionBuilder contracttransactionbuilder.Port,
	transactionRequest *contracttransactionbuilder.BuildTransactionRequest,
	transactionResponse *portcommon.TransactionResponse,
) (*portcommon.TransactionResponse, error) {
	if network.GasLimitMultiplier.LessThan(*amount.One()) {
		return transactionResponse, nil
	}

	transactionEstimationSender, err := s.transactionGasEstimationRegistry.GetTransactionGasEstimator(dlt)
	if err != nil {
		return nil, err
	}

	estimationResponse, err := transactionEstimationSender.EstimateGas(
		ctx,
		transactiongasestimator.EstimationRequest{
			From:        from,
			NetworkId:   network.Id,
			Transaction: transactionResponse,
		})
	if err != nil {
		return nil, err
	}

	multiplied := estimationResponse.Estimation.MulWithDecimals(network.GasLimitMultiplier, 0)
	multipliedGasLimit := &multiplied

	return transactionBuilder.OverrideGasLimit(*transactionRequest, *transactionResponse, multipliedGasLimit)
}

func (s *AppService) buildEvmTransactionRequest(
	ctx context.Context,
	req TransactionRequest,
	network config.NetworkConfig,
) (*contracttransactionbuilder.BuildTransactionRequest, error) {
	nonce, err := s.nonceProvider.GetNonce(ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    req.NetworkId,
		DltAccountId: req.SenderDltAccountId,
	})
	if err != nil {
		return nil, err
	}

	evmClient, err := s.evmClientRegistry.GetClientForNetworkId(ctx, network.Id)
	if err != nil {
		return nil, err
	}

	switch network.TransactionType {
	case portcommon.TransactionTypeLegacy:
		gasPrice := s.resolveEvmGasPrice(ctx, evmClient, &network)
		return contracttransactionbuilder.NewEVMLegacyBuildTransactionRequest(
			req.SenderDltAccountId,
			req.SmartContractId,
			req.MethodName,
			req.MethodArgs,
			&network.ChainId,
			nonce,
			&network.GasLimit,
			gasPrice,
			amount.Zero(),
		), nil
	case portcommon.TransactionTypeDynamicFee:
		maxFeePerGas, err := s.resolveEvmMaxFeePerGas(ctx, evmClient, &network)
		if err != nil {
			return nil, err
		}
		return contracttransactionbuilder.NewEVMDynamicFeeBuildTransactionRequest(
			req.SenderDltAccountId,
			req.SmartContractId,
			req.MethodName,
			req.MethodArgs,
			&network.ChainId,
			nonce,
			&network.GasLimit,
			&network.MaxPriorityFeePerGas,
			maxFeePerGas,
			amount.Zero(),
		), nil
	default:
		return nil, domainerrors.NewInvalidTransactionTypeDomainError(network.TransactionType)
	}
}

func (s *AppService) buildSvmTransactionRequest(
	ctx context.Context,
	req TransactionRequest,
	network config.NetworkConfig,
) (*contracttransactionbuilder.BuildTransactionRequest, error) {
	blockhash, err := s.blockhashProvider.GetRecentBlockhash(ctx, network.Id)
	if err != nil {
		return nil, err
	}

	cuPrice, err := s.resolveSvmCuPrice(ctx, &network)
	if err != nil {
		return nil, err
	}

	return contracttransactionbuilder.NewSVMBuildTransactionRequest(
		req.SenderDltAccountId,
		req.SmartContractId,
		req.MethodName,
		req.MethodArgs,
		blockhash,
		cuPrice,
		&network.GasLimit,
	), nil
}

func (s *AppService) resolveEvmGasPrice(ctx context.Context, evmClient evm.Client, network *config.NetworkConfig) *amount.Amount {
	gasPrice, err := evmClient.GetGasPrice(ctx)
	if err != nil {
		logger.WarnWithCtx(ctx, "error fetching gas price, using configured default", "error", err)
		return &network.GasPrice
	}

	return gasPrice
}

func (s *AppService) resolveEvmMaxFeePerGas(ctx context.Context, evmClient evm.Client, network *config.NetworkConfig) (*amount.Amount, error) {
	baseFeePerGas, err := evmClient.GetBaseFeePerGas(ctx)
	if err != nil {
		return nil, err
	}

	incrementedBaseFee := baseFeePerGas.MulWithDecimals(network.FeeMultiplier, 0)
	maxFeePerGas := incrementedBaseFee.Add(network.MaxPriorityFeePerGas)
	return &maxFeePerGas, nil
}

func (s *AppService) resolveSvmCuPrice(ctx context.Context, network *config.NetworkConfig) (*amount.Amount, error) {
	svmClient, err := s.svmClientRegistry.GetClientForNetworkId(network.Id)
	if err != nil {
		return nil, err
	}

	recentFees, err := svmClient.GetRecentPrioritizationFees(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(recentFees, func(i, j int) bool { return recentFees[i] < recentFees[j] })

	var p75Fee uint64
	if len(recentFees) > 0 {
		idx := (len(recentFees)*3 - 1) / 4
		p75Fee = recentFees[idx]
	}

	cuPrice, err := amount.New(new(big.Int).SetUint64(p75Fee), 0)
	if err != nil {
		return nil, err
	}

	if cuPrice.GreaterThan(network.MaxCuPrice) {
		return &network.MaxCuPrice, nil
	}

	return cuPrice, nil
}
