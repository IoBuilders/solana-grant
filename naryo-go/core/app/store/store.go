package store

import (
	"context"
	"reflect"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

// Store persists values of type D.
//
// It takes no store configuration: the connection is handed to the adapter at
// construction, and which Node an item belongs to comes from the item itself.
// The configuration is what picks a Store — through Supports — not what drives
// one.
type Store[D any] interface {
	Save(ctx context.Context, data D) error

	Supports(storeType store.Type, dataType reflect.Type) bool
}
