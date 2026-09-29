package event

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type StoreWorker struct {
	interval   time.Duration
	batchSize  int
	repository eventstorerepo.Repository
	relay      Relay
	running    atomic.Bool
	isCross    bool
}

func NewStoreWorker(interval time.Duration, batchSize int, repository eventstorerepo.Repository, relay Relay, isCross bool) *StoreWorker {
	return &StoreWorker{
		interval:   interval,
		batchSize:  batchSize,
		repository: repository,
		relay:      relay,
		isCross:    isCross,
	}
}

func (r *StoreWorker) Start(ctx context.Context) error {
	defer r.running.Store(false)
	if r.running.Swap(true) {
		logger.WarnWithCtx(ctx, "StoreWorker is already running, ignoring call")
		return nil
	}
	return r.run(ctx)
}

func (r *StoreWorker) run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil {
				logger.ErrorWithCtx(ctx, "Error processing event batch", "error", err)
			}
		}
	}
}

func (r *StoreWorker) processBatch(ctx context.Context) error {
	events, err := r.repository.ClaimPendingEvents(ctx, r.isCross, r.batchSize)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		logger.DebugWithCtx(ctx, fmt.Sprintf("Processing %d events", len(events)))
		if err := r.relay.Relay(ctx, &events); err != nil {
			return err
		}
	}
	return nil
}
