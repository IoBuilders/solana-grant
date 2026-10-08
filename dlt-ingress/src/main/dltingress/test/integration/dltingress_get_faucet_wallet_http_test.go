//go:build test || integration

package integration

import (
	"net/http"
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/infra/api/getfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/api/model"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"

	"github.com/stretchr/testify/require"
)

// The success path needs a real balance lookup, which this harness can't provide (see overrideDltIngressNetwork in main_test.go); it's covered by the mocked unit tests instead.
func TestDltIngressGetFaucetWallet(t *testing.T) {
	t.Run("Get faucet wallet for unknown network returns 404", func(t *testing.T) {
		_, errorResponse, resp, _ := getFaucetWallet("unknown-network")

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		require.Equal(t, "NETWORK_NOT_FOUND", errorResponse.Code)
	})
}

func getFaucetWallet(networkId string) (model.FaucetWalletBalanceModel, api.ErrorResponse, *http.Response, error) {
	path := strings.Replace(getfaucetwallet.UrlPath, ":networkId", networkId, 1)
	return HttpCallJSONResponse[model.FaucetWalletBalanceModel](
		http.MethodGet,
		"/api/v1"+path,
		nil,
		http.StatusOK,
	)
}
