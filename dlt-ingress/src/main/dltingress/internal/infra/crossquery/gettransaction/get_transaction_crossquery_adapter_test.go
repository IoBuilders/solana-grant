package gettransactionadapter

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/query"
	gettransactioncross "dlt-ingress/src/main/dltingress/port/crossquery/gettransaction"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/app/query/gettransaction"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"

	querybusmock "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/query"
)

func TestDltIngressGetTransactionCrossQueryAdapter_Execute_Success(t *testing.T) {
	queryBus := new(querybusmock.QueryBusMock)

	expectedResp := gettransaction.Response{
		TxId:      "0xabc123",
		NetworkId: "1",
		Dlt:       "EVM",
		Evm:       &query.EvmTransaction{FromAddress: "0xFrom"},
	}

	queryBus.On("Dispatch", mock.Anything, mock.MatchedBy(func(query gettransaction.Query) bool {
		return query.TxId == "0xabc123" && query.Dlt == "EVM"
	})).Return(expectedResp, nil)

	adapter := NewCrossQueryAdapter(queryBus)
	rawResp, err := adapter.Execute(context.Background(), gettransactioncross.CrossQuery{
		TxId: "0xabc123",
		Dlt:  "EVM",
	})

	assert.Nil(t, err)
	resp, ok := rawResp.(gettransactioncross.Response)
	assert.True(t, ok)
	assert.Equal(t, expectedResp.TxId, resp.TxId)
	assert.Equal(t, expectedResp.NetworkId, resp.NetworkId)
	assert.Equal(t, expectedResp.Dlt, resp.Dlt)
	assert.Equal(t, expectedResp.Evm.FromAddress, resp.Evm.FromAddress)
	queryBus.AssertExpectations(t)
}

func TestDltIngressGetTransactionCrossQueryAdapter_Execute_QueryBusError(t *testing.T) {
	queryBus := new(querybusmock.QueryBusMock)

	queryBus.On("Dispatch", mock.Anything, mock.Anything).
		Return(gettransaction.Response{}, domainerrors.NewEntityNotFoundDomainError("Transaction", uuid.Nil))

	adapter := NewCrossQueryAdapter(queryBus)
	rawResp, err := adapter.Execute(context.Background(), gettransactioncross.CrossQuery{
		TxId: "0xNotFound",
		Dlt:  "EVM",
	})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	queryBus.AssertExpectations(t)
}
