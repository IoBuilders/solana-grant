package getfailedevents

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gorm.io/datatypes"
)

func TestCoreGetFailedeventsQueryHandler_Execute(t *testing.T) {
	type testCase struct {
		name             string
		expectedItems    []*EventConsumerQueryResponse
		setupMocks       func()
		shouldFail       bool
		expectedErrorMsg string
	}

	ec1 := &eventstore.EventConsumer{
		Model:  basemodel.Model{Id: uuid.New(), CreatedAt: time.Now()},
		Status: eventstore.Failed,
		EventStore: eventstore.EventStore{
			Type:    "EventType1",
			Payload: datatypes.JSON("{}"),
		},
		ErrorDetails: new("test error details"),
	}
	ec2 := &eventstore.EventConsumer{
		Model:  basemodel.Model{Id: uuid.New(), CreatedAt: time.Now()},
		Status: eventstore.Failed,
		EventStore: eventstore.EventStore{
			Type:    "EventType2",
			Payload: datatypes.JSON("{}"),
		},
		ErrorDetails: new("test error details 2"),
	}

	domainEventConsumers := []eventstore.EventConsumer{*ec1, *ec2}
	expectedItems := []*EventConsumerQueryResponse{toItem(*ec1), toItem(*ec2)}

	var repo *eventstorerepo.EventConsumerRepositoryMock

	cases := []testCase{
		{
			name:          "Success - Empty array",
			expectedItems: []*EventConsumerQueryResponse{},
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, eventstore.Failed, mock.Anything).
					Return([]eventstore.EventConsumer{}, nil).Once()
				repo.On("CountAllByFilters",
					mock.Anything, eventstore.Failed).
					Return(0, nil).Once()
			},
		},
		{
			name:          "Success",
			expectedItems: expectedItems,
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, eventstore.Failed, mock.Anything).
					Return(domainEventConsumers, nil).Once()
				repo.On("CountAllByFilters",
					mock.Anything, eventstore.Failed).
					Return(len(domainEventConsumers), nil).Once()
			},
		},
		{
			name: "Error - Repository Find Failure",
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, eventstore.Failed, mock.Anything).
					Return([]eventstore.EventConsumer{}, fmt.Errorf("db error")).Once()
			},
			shouldFail:       true,
			expectedErrorMsg: "db error",
		},
		{
			name: "Error - Repository Count Failure",
			setupMocks: func() {
				repo.On("FindByStatusPaginated",
					mock.Anything, eventstore.Failed, mock.Anything).
					Return([]eventstore.EventConsumer{}, nil).Once()
				repo.On("CountAllByFilters",
					mock.Anything, eventstore.Failed).
					Return(0, fmt.Errorf("db error")).Once()
			},
			shouldFail:       true,
			expectedErrorMsg: "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo = new(eventstorerepo.EventConsumerRepositoryMock)
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
