//go:build test || integration

package integration

import (
	"net/http"
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/api/createfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/api/model"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/stretchr/testify/require"
)

func TestDltIngressCreateFaucetWallet(t *testing.T) {
	t.Run("Create faucet wallet for unknown network returns 404", func(t *testing.T) {
		fundingAmount, _ := amount.NewFromString("1000000000000000000")
		balanceThreshold, _ := amount.NewFromString("100000000000000000")

		_, errorResponse, resp, _ := createFaucetWallet("unknown-network", createfaucetwallet.Request{
			FundingAmount:    *fundingAmount,
			BalanceThreshold: *balanceThreshold,
		})

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		require.Equal(t, "NETWORK_NOT_FOUND", errorResponse.Code)
	})

	t.Run("Create faucet wallet with balance threshold not lower than funding amount returns 400", func(t *testing.T) {
		fundingAmount, _ := amount.NewFromString("1000000000000000000")

		_, errorResponse, resp, _ := createFaucetWallet("default", createfaucetwallet.Request{
			FundingAmount:    *fundingAmount,
			BalanceThreshold: *fundingAmount,
		})

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.Equal(t, string(domainerrors.ErrorCodeInvalidBalanceThreshold), errorResponse.Code)
	})

	t.Run("Create faucet wallet success", func(t *testing.T) {
		fundingAmount, _ := amount.NewFromString("1000000000000000000")
		balanceThreshold, _ := amount.NewFromString("100000000000000000")

		response, _, resp, err := createFaucetWallet("default", createfaucetwallet.Request{
			FundingAmount:    *fundingAmount,
			BalanceThreshold: *balanceThreshold,
		})

		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "default", response.NetworkId)
		require.NotEmpty(t, response.Id)
		require.NotEmpty(t, response.DltAccountId)
		require.True(t, fundingAmount.Equal(response.FundingAmount))
		require.True(t, balanceThreshold.Equal(response.BalanceThreshold))
		require.True(t, response.Enabled)

		exists, err := DltIngress.Repositories.FaucetWalletRepo.ExistByNetworkId(ctx, "default")
		require.Nil(t, err)
		require.True(t, exists)

		wallet, err := DltIngress.Repositories.FaucetWalletRepo.FindByNetworkId(ctx, "default")
		require.Nil(t, err)
		require.Equal(t, response.Id, wallet.Id.String())
		require.True(t, wallet.Enabled)

		custodyKey, err := DltIngress.Repositories.CustodyKeyRepo.FindById(ctx, wallet.CustodyKeyId)
		require.Nil(t, err)
		require.Equal(t, response.DltAccountId, custodyKey.DltAccountId)

		t.Run("Create faucet wallet again for the same network returns 400", func(t *testing.T) {
			_, errorResponse, resp, _ := createFaucetWallet("default", createfaucetwallet.Request{
				FundingAmount:    *fundingAmount,
				BalanceThreshold: *balanceThreshold,
			})

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			require.Equal(t, string(domainerrors.ErrorCodeFaucetWalletAlreadyExists), errorResponse.Code)
		})
	})
}

func createFaucetWallet(networkId string, request createfaucetwallet.Request) (model.FaucetWalletModel, api.ErrorResponse, *http.Response, error) {
	path := strings.Replace(createfaucetwallet.UrlPath, ":networkId", networkId, 1)
	return HttpCallJSONResponse[model.FaucetWalletModel](
		http.MethodPost,
		"/api/v1"+path,
		request,
		http.StatusOK,
	)
}
