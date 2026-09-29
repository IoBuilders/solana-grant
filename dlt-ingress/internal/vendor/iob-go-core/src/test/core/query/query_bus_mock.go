package querybusmock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

type QueryBusMock struct {
	mock.Mock
}

func (m *QueryBusMock) Register(handler interface{}) error {
	m.Called(handler)
	return nil
}

func (m *QueryBusMock) RegisterHandle(handler any) error {
	m.Called(handler)
	return nil
}

func (m *QueryBusMock) Dispatch(ctx context.Context, queryRequest query.Query) (query.Response, error) {
	args := m.Called(ctx, queryRequest)
	if args.Get(0) != nil {
		return args.Get(0).(query.Response), args.Error(1)
	}
	return nil, args.Error(1)
}

var _ query.Bus = (*QueryBusMock)(nil)
