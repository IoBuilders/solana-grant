package blockprocessor

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/dispatch"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// ContractEventDispatcherHelper builds a SolanaContractEvent from decoded
// filter-match data and dispatches it. It exists so that whichever trigger
// eventually decodes a matched transaction (e.g. a future
// SolanaTransactionProcessorPermanentTrigger) does not need to construct and
// dispatch SolanaContractEvent values itself.
type ContractEventDispatcherHelper struct {
	dispatcher dispatch.Dispatcher
}

// NewContractEventDispatcherHelper builds a ContractEventDispatcherHelper
// backed by dispatcher.
func NewContractEventDispatcherHelper(dispatcher dispatch.Dispatcher) *ContractEventDispatcherHelper {
	return &ContractEventDispatcherHelper{dispatcher: dispatcher}
}

// Execute builds a SolanaContractEvent from its constituent decoded data and
// dispatches it.
func (h *ContractEventDispatcherHelper) Execute(
	ctx context.Context,
	nodeID uuid.UUID,
	programID, signature string,
	slot uint64,
	parameters []parameter.ContractEventParameter,
	eventName string,
	status event.ContractEventStatus,
) error {
	contractEvent, err := event.NewSolanaContractEvent(nodeID, programID, signature, slot, parameters, eventName, status)
	if err != nil {
		return fmt.Errorf("contract event dispatcher helper: failed to build contract event: %w", err)
	}

	h.dispatcher.Dispatch(ctx, contractEvent)
	return nil
}
