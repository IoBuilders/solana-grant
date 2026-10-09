package getfailedtransactions

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/query"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/failedtransaction/mock"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

func TestDltIngressGetFailedTransactionsQueryHandler_Execute(t *testing.T) {
	type testCase struct {
		name             string
		expectedItems    []*query.FailedTransaction
		setupMocks       func()
		shouldFail       bool
		expectedErrorMsg string
	}

	ec1 := &failedtransaction.FailedTransaction{
		Entity:       base.Entity{Id: uuid.New(), CreatedAt: time.Now()},
		Status:       failedtransaction.StatusNotRetried,
		ErrorDetails: "test error details",
	}
	ec2 := &failedtransaction.FailedTransaction{
		Entity:       base.Entity{Id: uuid.New(), CreatedAt: time.Now()},
		Status:       failedtransaction.StatusNotRetried,
		ErrorDetails: "test error details 2",
	}

	domainFailedTransactions := []failedtransaction.FailedTransaction{*ec1, *ec2}
	expectedItems := []*query.FailedTransaction{ToQueryResponse(*ec1), ToQueryResponse(*ec2)}

	var repo *mockfailedtransactionrepo.Postgres

	cases := []testCase{
		{
			name:          "Success - Empty array",
			expectedItems: []*query.FailedTransaction{},
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, failedtransaction.StatusNotRetried, mock.Anything).
					Return([]failedtransaction.FailedTransaction{}, nil).Once()
				repo.On("CountAllByFilters",
					mock.Anything, failedtransaction.StatusNotRetried).
					Return(0, nil).Once()
			},
		},
		{
			name:          "Success",
			expectedItems: expectedItems,
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, failedtransaction.StatusNotRetried, mock.Anything).
					Return(domainFailedTransactions, nil).Once()
				repo.On("CountAllByFilters",
					mock.Anything, failedtransaction.StatusNotRetried).
					Return(len(domainFailedTransactions), nil).Once()
			},
		},
		{
			name: "Error - Repository Find Failure",
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, failedtransaction.StatusNotRetried, mock.Anything).
					Return([]failedtransaction.FailedTransaction{}, fmt.Errorf("db error")).Once()
			},
			shouldFail:       true,
			expectedErrorMsg: "db error",
		},
		{
			name: "Error - Repository Count Failure",
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, failedtransaction.StatusNotRetried, mock.Anything).
					Return([]failedtransaction.FailedTransaction{}, nil).Once()
				repo.On("CountAllByFilters",
					mock.Anything, failedtransaction.StatusNotRetried).
					Return(0, fmt.Errorf("db error")).Once()
			},
			shouldFail:       true,
			expectedErrorMsg: "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo = new(mockfailedtransactionrepo.Postgres)
			tc.setupMocks()

			handler := NewHandler(repo)

			resp, err := handler.Execute(context.Background(), Query{
				PaginationParams: pagination.PaginationParams{PageSize: 10, Offset: 0},
			})

			if !tc.shouldFail {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedItems, resp.Items)
				assert.Equal(t, len(tc.expectedItems), resp.TotalElements)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErrorMsg)
			}
			repo.AssertExpectations(t)
		})
	}
}

func ToQueryResponse(ec failedtransaction.FailedTransaction) *query.FailedTransaction {
	return &query.FailedTransaction{
		TxId:         ec.TxId,
		NetworkId:    ec.NetworkId,
		CreatedAt:    ec.CreatedAt,
		ErrorDetails: ec.ErrorDetails,
	}
}
