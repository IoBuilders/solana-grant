package subscription

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/dispatch"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/error"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/retry"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

type SolanaPollSlotSubscriber struct {
	slot                uint64
	n                   *node.Node
	blockInteractor     interactor.BlockInteractor
	dispatcher          dispatch.Dispatcher
	startSlotCalculator *SolanaStartSlotCalculator
	retryer             retry.Retryer
	retryConfiguration  *common.RetryConfiguration
}

func NewSolanaPollSlotSubscriber(n *node.Node,
	blockInteractor interactor.BlockInteractor,
	dispatcher dispatch.Dispatcher,
	startSlotCalculator *SolanaStartSlotCalculator,
	retryer retry.Retryer,
	retryConfiguration *common.RetryConfiguration,
) *SolanaPollSlotSubscriber {
	return &SolanaPollSlotSubscriber{
		slot:                0,
		n:                   n,
		blockInteractor:     blockInteractor,
		dispatcher:          dispatcher,
		startSlotCalculator: startSlotCalculator,
		retryer:             retryer,
		retryConfiguration:  retryConfiguration,
	}
}

func (s *SolanaPollSlotSubscriber) Subscribe(ctx context.Context) {
	go func() {
		defer s.recoverPanic(ctx)
		logging.InfoWithCtx(ctx, "Subscription started", "node", s.n.Name)
		s.handleSlots(ctx)
	}()
}

func (s *SolanaPollSlotSubscriber) handleSlots(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logging.InfoWithCtx(ctx, "Subscription stopped", "node", s.n.Name)
			return
		default:
			_, err := s.retryer.Execute(ctx, retry.FromRetryConfiguration(s.retryConfiguration, nil, nil), func(ctx context.Context) (any, error) {
				if s.slot == 0 {
					// First iteration, get starting point
					slot, err := s.startSlotCalculator.GetStartSlot(ctx)
					if err != nil {
						return nil, err
					}
					s.slot = slot
				} else {
					// Check if slot is available, if not, wait polling time
					currentSlot, err := s.blockInteractor.GetSlot(ctx)
					if err != nil {
						return nil, err
					}
					if s.slot > currentSlot {
						select {
						case <-ctx.Done():
						case <-time.After(s.n.Subscription.MethodConfiguration.(node.PollBlockSubscriptionMethodConfiguration).Interval):
						}
						return nil, nil
					}
				}
				logging.DebugWithCtx(ctx, fmt.Sprintf("Processing slot: %d", s.slot), "node", s.n.Name)

				// Dispatch event and increase slot number
				slotEvent, _ := event.NewSlotEvent(s.n.ID, s.slot, time.Now())
				s.dispatcher.Dispatch(ctx, slotEvent)
				s.slot++
				return nil, nil
			})
			if err != nil {
				logging.ErrorWithCtx(ctx,
					"Subscription stopped due to error after retries",
					"error", err,
					"slot", s.slot,
				)
				return
			}
		}
	}
}

func (s *SolanaPollSlotSubscriber) recoverPanic(ctx context.Context) {
	if r := recover(); r != nil {
		info := panicinfo.NewPanicInfo(r)
		logging.ErrorWithCtx(ctx,
			fmt.Sprintf("%s %s", panicinfo.PANIC_RECOVERED_TAG, "Subscription panic recovered. Subscription stopped"),
			"slot", s.slot,
			"node", s.n.Name,
			"panic", info.Message,
			"file", info.File,
			"line", info.Line,
			"trace", info.FormatStackTrace(),
		)
	}
}
