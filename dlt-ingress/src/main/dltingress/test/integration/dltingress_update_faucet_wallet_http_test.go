//go:build test || integration

package integration

import (
	"net/http"
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/api/createfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/api/model"
	"dlt-ingress/src/main/dltingress/internal/infra/api/updatefaucetwallet"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/stretchr/testify/require"
)

func TestDltIngressUpdateFaucetWallet(t *testing.T) {
	ensureFaucetWalletExists(t, "default")

	t.Run("Update faucet wallet for unknown network returns 404", func(t *testing.T) {
		fundingAmount, _ := amount.NewFromString("2000000000000000000")
		balanceThreshold, _ := amount.NewFromString("200000000000000000")
		enabled := true

		_, errorResponse, resp, _ := updateFaucetWallet("unknown-network", updatefaucetwallet.Request{
			FundingAmount:    *fundingAmount,
			BalanceThreshold: *balanceThreshold,
			Enabled:          &enabled,
		})

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		require.Equal(t, string(domainerrors.ErrorCodeFaucetWalletNotFoundByNetworkId), errorResponse.Code)
	})

	t.Run("Update faucet wallet with balance threshold not lower than funding amount returns 400", func(t *testing.T) {
		fundingAmount, _ := amount.NewFromString("2000000000000000000")
		enabled := true

		_, errorResponse, resp, _ := updateFaucetWallet("default", updatefaucetwallet.Request{
			FundingAmount:    *fundingAmount,
			BalanceThreshold: *fundingAmount,
			Enabled:          &enabled,
		})

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.Equal(t, string(domainerrors.ErrorCodeInvalidBalanceThreshold), errorResponse.Code)
	})

	t.Run("Update faucet wallet without enabled returns 400", func(t *testing.T) {
		fundingAmount, _ := amount.NewFromString("2000000000000000000")
		balanceThreshold, _ := amount.NewFromString("200000000000000000")

		_, _, resp, _ := updateFaucetWallet("default", updatefaucetwallet.Request{
			FundingAmount:    *fundingAmount,
			BalanceThreshold: *balanceThreshold,
		})

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Update faucet wallet success", func(t *testing.T) {
		fundingAmount, _ := amount.NewFromString("2000000000000000000")
		balanceThreshold, _ := amount.NewFromString("200000000000000000")
		enabled := false

		response, _, resp, err := updateFaucetWallet("default", updatefaucetwallet.Request{
			FundingAmount:    *fundingAmount,
			BalanceThreshold: *balanceThreshold,
			Enabled:          &enabled,
		})

		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "default", response.NetworkId)
		require.NotEmpty(t, response.Id)
		require.True(t, fundingAmount.Equal(response.FundingAmount))
		require.True(t, balanceThreshold.Equal(response.BalanceThreshold))
		require.Equal(t, enabled, response.Enabled)

		wallet, err := DltIngress.Repositories.FaucetWalletRepo.FindByNetworkId(ctx, "default")
		require.Nil(t, err)
		require.Equal(t, response.Id, wallet.Id.String())
		require.True(t, fundingAmount.Equal(wallet.FundingAmount))
		require.True(t, balanceThreshold.Equal(wallet.BalanceThreshold))
		require.Equal(t, enabled, wallet.Enabled)
	})
}

func ensureFaucetWalletExists(t *testing.T, networkId string) {
	exists, err := DltIngress.Repositories.FaucetWalletRepo.ExistByNetworkId(ctx, networkId)
	require.Nil(t, err)
	if exists {
		return
	}

	fundingAmount, _ := amount.NewFromString("1000000000000000000")
	balanceThreshold, _ := amount.NewFromString("100000000000000000")

	_, _, resp, err := createFaucetWallet(networkId, createfaucetwallet.Request{
		FundingAmount:    *fundingAmount,
		BalanceThreshold: *balanceThreshold,
	})
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func updateFaucetWallet(networkId string, request updatefaucetwallet.Request) (model.FaucetWalletModel, api.ErrorResponse, *http.Response, error) {
	path := strings.Replace(updatefaucetwallet.UrlPath, ":networkId", networkId, 1)
	return HttpCallJSONResponse[model.FaucetWalletModel](
		http.MethodPut,
		"/api/v1"+path,
		request,
		http.StatusOK,
	)
}
