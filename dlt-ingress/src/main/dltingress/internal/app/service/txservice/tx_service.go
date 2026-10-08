package txservice

import (
	"context"
	"errors"
	"maps"
	"math/big"
	"slices"
	"sort"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

const NativeTransferMethodName = "nativeTransfer"

type TransactionRequest struct {
	SenderDltAccountId   string
	SignersDltAccountIds []string
	SmartContractId      string
	SmartContractName    string
	MethodName           string
	MethodArgs           map[string]any
	NetworkId            string
	ResolveNestedCalls   bool
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

	var transactionBuilder contracttransactionbuilder.Port
	if req.MethodName == NativeTransferMethodName {
		transactionBuilder = s.transactionBuilderRegistry.GetNativeTransferTransactionBuilder()
	} else {
		transactionBuilder, err = s.transactionBuilderRegistry.GetContractTransactionBuilder(dlt, req.SmartContractName)
		if err != nil {
			return nil, err
		}
	}

	if req.ResolveNestedCalls {
		req, err = s.resolveNestedCalls(dlt, req)
		if err != nil {
			return nil, err
		}
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

	if dlt == common.SVM {
		transactionRequest, transaction, err = s.applySvmCuPrice(ctx, &network, transactionBuilder, transactionRequest, transaction)
		if err != nil {
			return nil, err
		}
	}

	transactionWithGasOverridden, err := s.estimateGasAndOverrideLimit(ctx, dlt, req.SenderDltAccountId, network, transactionBuilder, transactionRequest, transaction)
	if err != nil {
		return nil, err
	}

	return transactionWithGasOverridden, nil
}

func (s *AppService) resolveNestedCalls(dlt common.Dlt, req TransactionRequest) (TransactionRequest, error) {
	if !req.ResolveNestedCalls {
		return req, nil
	}

	var resolved map[string]any

	for argName, value := range req.MethodArgs {
		calls, isNested := value.([]portcommon.Invocation)
		if !isNested {
			continue
		}

		encodedCalls := make([][]byte, 0, len(calls))
		for i, call := range calls {
			if containsCalls(call.MethodArgs) {
				return req, domainerrors.NewNestedCallDepthDomainError(argName, i, call.SmartContractName, call.MethodName)
			}

			builder, err := s.transactionBuilderRegistry.GetContractTransactionBuilder(dlt, call.SmartContractName)
			if err != nil {
				return req, err
			}
			encoded, err := builder.EncodeCallData(call.MethodName, call.MethodArgs)
			if err != nil {
				return req, domainerrors.NewNestedCallEncodingDomainError(argName, i, call.SmartContractName, call.MethodName, err)
			}
			encodedCalls = append(encodedCalls, encoded)
		}
		if resolved == nil {
			resolved = make(map[string]any, len(req.MethodArgs))
			maps.Copy(resolved, req.MethodArgs)
		}
		resolved[argName] = encodedCalls
	}

	if resolved != nil {
		req.MethodArgs = resolved
	}

	return req, nil
}

func containsCalls(value any) bool {
	switch typed := value.(type) {
	case []portcommon.Invocation:
		return true
	case map[string]any:
		for _, v := range typed {
			if containsCalls(v) {
				return true
			}
		}
	case []any:
		return slices.ContainsFunc(typed, containsCalls)
	}
	return false
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
		logger.WarnWithCtx(ctx, "no gas estimator available", "error", err)
		return transactionResponse, nil
	}

	estimationResponse, err := transactionEstimationSender.EstimateGas(
		ctx,
		transactiongasestimator.EstimationRequest{
			From:        from,
			NetworkId:   network.Id,
			Transaction: transactionResponse,
		})

	if err != nil && dlt == common.SVM && errors.Is(err, svm.ErrBlockhashNotFound) {
		// The blockhash this transaction was built with is already gone (the cache served a stale
		// value, or it just expired between fetch and simulation). Force a fresh one and rebuild
		// once before falling back to the no-estimate path below.
		if rebuilt, rebuildErr := s.refreshSvmBlockhashAndRebuild(ctx, network, transactionBuilder, transactionRequest); rebuildErr != nil {
			logger.WarnWithCtx(ctx, "error refetching blockhash after stale-blockhash simulation failure", "error", rebuildErr)
		} else {
			transactionResponse = rebuilt
			estimationResponse, err = transactionEstimationSender.EstimateGas(
				ctx,
				transactiongasestimator.EstimationRequest{
					From:        from,
					NetworkId:   network.Id,
					Transaction: transactionResponse,
				})
		}
	}

	if err != nil {
		logger.WarnWithCtx(ctx, "error estimating gas", "error", err)
		if dlt == common.SVM {
			// A failed simulation leaves no reliable estimate to size the compute unit limit with, and the
			// seed used to build this transaction is sized to never clip a legitimate simulation, so it is
			// not a safe value to submit as-is either. Zero (rather than nil) clears the limit so the
			// cluster's own per-instruction default applies: CuLimit is persisted to a NOT NULL column,
			// and the builder already treats a zero amount the same as an absent one.
			return transactionBuilder.OverrideGasLimit(*transactionRequest, *transactionResponse, amount.Zero())
		}
		return transactionResponse, nil
	}

	if s.exceedsGasCap(dlt, network, *estimationResponse.Estimation) {
		return nil, domainerrors.NewGasLimitExceedsMaximumDomainError(*estimationResponse.Estimation, s.gasCap(dlt, network))
	}

	multiplied := estimationResponse.Estimation.MulWithDecimals(network.GasLimitMultiplier, 0)

	if s.exceedsGasCap(dlt, network, multiplied) {
		return transactionResponse, nil
	}

	return transactionBuilder.OverrideGasLimit(*transactionRequest, *transactionResponse, &multiplied)
}

func (s *AppService) exceedsGasCap(dlt common.Dlt, network config.NetworkConfig, value amount.Amount) bool {
	switch dlt {
	case common.EVM:
		return network.HasGasCap() && value.GreaterThan(network.GasLimit)
	case common.SVM:
		return network.HasCuLimitCap() && value.GreaterThan(network.MaxCuLimit)
	default:
		return false
	}
}

func (s *AppService) gasCap(dlt common.Dlt, network config.NetworkConfig) amount.Amount {
	if dlt == common.SVM {
		return network.MaxCuLimit
	}
	return network.GasLimit
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

	// CuPrice is left unset: the accounts the fee estimate depends on are only known once the transaction
	// has been built, so applySvmCuPrice resolves and applies it afterwards.
	return contracttransactionbuilder.NewSVMBuildTransactionRequest(
		req.SenderDltAccountId,
		req.SmartContractId,
		req.MethodName,
		req.MethodArgs,
		blockhash,
		nil,
		&network.MaxCuLimit,
	), nil
}

// refreshSvmBlockhashAndRebuild discards the cached blockhash, fetches a fresh one, and rebuilds the
// transaction with it. transactionRequest is updated in place so every subsequent step (gas cap check,
// OverrideGasLimit) keeps seeing the blockhash that was actually simulated and will be signed.
func (s *AppService) refreshSvmBlockhashAndRebuild(
	ctx context.Context,
	network config.NetworkConfig,
	transactionBuilder contracttransactionbuilder.Port,
	transactionRequest *contracttransactionbuilder.BuildTransactionRequest,
) (*portcommon.TransactionResponse, error) {
	if err := s.blockhashProvider.InvalidateRecentBlockhash(ctx, network.Id); err != nil {
		return nil, err
	}

	blockhash, err := s.blockhashProvider.GetRecentBlockhash(ctx, network.Id)
	if err != nil {
		return nil, err
	}

	inner := *transactionRequest.SVMBuildTransactionRequest
	inner.RecentBlockHash = blockhash
	transactionRequest.SVMBuildTransactionRequest = &inner

	return transactionBuilder.BuildTransaction(transactionRequest)
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

// applySvmCuPrice prices the transaction from the recent prioritization fees paid on the accounts it write-locks,
// so the estimate reflects local contention rather than overall ledger activity. The price is set on the returned
// request as well as on the rebuilt transaction, because later overrides (OverrideGasLimit) rebuild from the request.
func (s *AppService) applySvmCuPrice(
	ctx context.Context,
	network *config.NetworkConfig,
	transactionBuilder contracttransactionbuilder.Port,
	transactionRequest *contracttransactionbuilder.BuildTransactionRequest,
	transactionResponse *portcommon.TransactionResponse,
) (*contracttransactionbuilder.BuildTransactionRequest, *portcommon.TransactionResponse, error) {
	cuPrice, err := s.resolveSvmCuPrice(ctx, network, transactionResponse.SVMTransactionResponse.WritableAccountKeys())
	if err != nil {
		return nil, nil, err
	}

	pricedRequest := *transactionRequest
	inner := *pricedRequest.SVMBuildTransactionRequest
	inner.CuPrice = cuPrice
	pricedRequest.SVMBuildTransactionRequest = &inner

	pricedResponse, err := transactionBuilder.OverrideCuPrice(pricedRequest, *transactionResponse, cuPrice)
	if err != nil {
		return nil, nil, err
	}

	return &pricedRequest, pricedResponse, nil
}

func (s *AppService) resolveSvmCuPrice(ctx context.Context, network *config.NetworkConfig, writableAccounts []string) (*amount.Amount, error) {
	svmClient, err := s.svmClientRegistry.GetClientForNetworkId(network.Id)
	if err != nil {
		return nil, err
	}

	recentFees, err := svmClient.GetRecentPrioritizationFees(ctx, writableAccounts)
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
