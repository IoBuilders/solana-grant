package fundaccount

import (
	"context"
	"errors"

	"dlt-ingress/src/main/dltingress/internal/app/service/txservice"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

type AppServiceInterface interface {
	Execute(ctx context.Context, request *Request) error
}

type AppService struct {
	faucetWalletRepo  faucetwallet.Repository
	custodyKeyRepo    custodykey.Repository
	evmClientRegistry evm.ClientRegistry
	svmClientRegistry svm.ClientRegistry
	commandBus        command.Bus
}

func NewAppService(
	faucetWalletRepo faucetwallet.Repository,
	custodyKeyRepo custodykey.Repository,
	evmClientRegistry evm.ClientRegistry,
	svmClientRegistry svm.ClientRegistry,
	commandBus command.Bus,
) *AppService {
	return &AppService{
		faucetWalletRepo:  faucetWalletRepo,
		custodyKeyRepo:    custodyKeyRepo,
		evmClientRegistry: evmClientRegistry,
		svmClientRegistry: svmClientRegistry,
		commandBus:        commandBus,
	}
}

func (s *AppService) Execute(ctx context.Context, request *Request) error {
	accountToFund, err := s.custodyKeyRepo.FindByDltAccountId(ctx, request.DltAccountId)
	if err != nil {
		return err
	}

	if request.NetworkId != nil {
		networkConfig, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(*request.NetworkId)
		if err != nil {
			return err
		}

		if err = networkConfig.CheckDltAccount(accountToFund.DltAccountId, string(accountToFund.Dlt)); err != nil {
			return err
		}

		return s.fundDltAccountForNetwork(ctx, networkConfig.Id, accountToFund)
	}

	networkConfigsForDlt := config.DltIngressConfig.DltIngress.GetDltIngressNetworksByDlt(string(accountToFund.Dlt))
	for _, networkConfig := range networkConfigsForDlt {
		err = s.fundDltAccountForNetwork(ctx, networkConfig.Id, accountToFund)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AppService) fundDltAccountForNetwork(ctx context.Context, networkId string, accountToFund *custodykey.CustodyKey) error {
	faucetWallet, err := s.faucetWalletRepo.FindByNetworkId(ctx, networkId)
	if err != nil {
		if errors.Is(err, coreerror.ErrNotFound) {
			return nil
		}
		return err
	}

	if !faucetWallet.Enabled {
		return nil
	}

	faucetWalletCustodyKey, err := s.custodyKeyRepo.FindById(ctx, faucetWallet.CustodyKeyId)
	if err != nil {
		return err
	}

	destinationWalletBalance, err := s.getBalance(ctx, accountToFund.Dlt, networkId, accountToFund.DltAccountId)
	if err != nil {
		return err
	}

	if destinationWalletBalance.LessThan(faucetWallet.BalanceThreshold) {
		amountToFund := faucetWallet.FundingAmount.Sub(*destinationWalletBalance)

		faucetWalletBalance, err := s.getBalance(ctx, accountToFund.Dlt, networkId, faucetWalletCustodyKey.DltAccountId)
		if err != nil {
			return err
		}

		if faucetWalletBalance.LessThan(amountToFund) {
			return domainerrors.NewFaucetWalletInsufficientFundsDomainError(networkId)
		}

		if _, err := s.commandBus.Dispatch(ctx, &signandsend.Command{
			SenderDltAccountId:   faucetWalletCustodyKey.DltAccountId,
			SignersDltAccountIds: []string{faucetWalletCustodyKey.DltAccountId},
			MethodName:           txservice.NativeTransferMethodName,
			MethodArgs: map[string]any{
				"to":     accountToFund.DltAccountId,
				"amount": amountToFund.String(),
			},
			NetworkId: networkId,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *AppService) getBalance(ctx context.Context, dlt common.Dlt, networkId, dltAccountId string) (*amount.Amount, error) {
	switch dlt {
	case common.EVM:
		client, err := s.evmClientRegistry.GetClientForNetworkId(ctx, networkId)
		if err != nil {
			return nil, err
		}
		return client.GetBalance(ctx, dltAccountId)
	case common.SVM:
		client, err := s.svmClientRegistry.GetClientForNetworkId(networkId)
		if err != nil {
			return nil, err
		}
		return client.GetBalance(ctx, dltAccountId)
	default:
		return nil, domainerrors.NewInvalidDltDomainError(string(dlt))
	}
}
