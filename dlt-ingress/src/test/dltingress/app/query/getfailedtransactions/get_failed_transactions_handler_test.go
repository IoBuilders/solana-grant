package getfailedtransactions

import (
	"context"
	"dlt-ingress/src/main/dltingress/app/query/getfailedtransactions"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	mocks "dlt-ingress/src/test/dltingress/port/repository"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
)

func TestDltIngressGetFailedTransactionsQueryHandler_Execute(t *testing.T) {
	type testCase struct {
		name             string
		expectedItems    []*getfailedtransactions.FailedTransactionQueryResponse
		setupMocks       func()
		shouldFail       bool
		expectedErrorMsg string
	}

	ec1 := &failedtransaction.FailedTransaction{
		Model:        basemodel.Model{Id: uuid.New(), CreatedAt: time.Now()},
		Status:       failedtransaction.StatusNotRetried,
		ErrorDetails: "test error details",
	}
	ec2 := &failedtransaction.FailedTransaction{
		Model:        basemodel.Model{Id: uuid.New(), CreatedAt: time.Now()},
		Status:       failedtransaction.StatusNotRetried,
		ErrorDetails: "test error details 2",
	}

	domainFailedTransactions := []failedtransaction.FailedTransaction{*ec1, *ec2}
	expectedItems := []*getfailedtransactions.FailedTransactionQueryResponse{getfailedtransactions.ToItem(*ec1), getfailedtransactions.ToItem(*ec2)}

	var repo *mocks.FailedTransactionRepositoryMock

	cases := []testCase{
		{
			name:          "Success - Empty array",
			expectedItems: []*getfailedtransactions.FailedTransactionQueryResponse{},
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
			repo = new(mocks.FailedTransactionRepositoryMock)
			tc.setupMocks()

			handler := getfailedtransactions.NewHandler(repo)

			resp, err := handler.Execute(context.Background(), getfailedtransactions.Query{
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
