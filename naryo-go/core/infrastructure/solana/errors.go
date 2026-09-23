package solana

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
)

// JSON-RPC custom error codes a Solana node's getBlock method can return.
// See https://github.com/anza-xyz/agave/blob/master/rpc-client-api/src/custom_error.rs.
const (
	errCodeBlockCleanedUp             = -32001
	errCodeBlockNotAvailable          = -32004
	errCodeSlotSkipped                = -32007
	errCodeLongTermStorageSlotSkipped = -32009
	errCodeBlockStatusNotAvailableYet = -32014
)

// mapRPCError maps a JSON-RPC error result from method to the block sentinel
// error it corresponds to, so callers can distinguish the cause with
// errors.Is. An error code not among those getBlock is known to return is
// wrapped as a plain error instead.
func mapRPCError(method string, e rpcError) error {
	switch e.Code {
	case errCodeSlotSkipped, errCodeLongTermStorageSlotSkipped:
		return fmt.Errorf("solana rpc: %s: %s: %w", method, e.Message, block.ErrSlotSkipped)
	case errCodeBlockNotAvailable, errCodeBlockStatusNotAvailableYet:
		return fmt.Errorf("solana rpc: %s: %s: %w", method, e.Message, block.ErrBlockNotAvailable)
	case errCodeBlockCleanedUp:
		return fmt.Errorf("solana rpc: %s: %s: %w", method, e.Message, block.ErrBlockCleanedUp)
	default:
		return fmt.Errorf("solana rpc: %s failed with code %d: %s", method, e.Code, e.Message)
	}
}
