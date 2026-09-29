package event

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/panicinfo"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type JanitorWorker struct {
	interval          time.Duration
	threshold         time.Duration
	eventConsumerRepo eventstorerepo.EventConsumerRepository
}

func NewJanitorWorker(eventConsumerRepo eventstorerepo.EventConsumerRepository, interval time.Duration, threshold time.Duration) *JanitorWorker {
	return &JanitorWorker{
		interval:          interval,
		threshold:         threshold,
		eventConsumerRepo: eventConsumerRepo,
	}
}

func (j *JanitorWorker) Start(ctx context.Context) {

	logger.InfoWithCtx(ctx, "[JANITOR_WORKER] Event consumer cleanup started with interval", "interval", j.interval)

	ticker := time.NewTicker(j.interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.InfoWithCtx(ctx, "[JANITOR_WORKER] Stopped")
				return
			case <-ticker.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							info := panicinfo.NewPanicInfo(r)
							logger.ErrorWithCtx(ctx,
								fmt.Sprintf("%s %s", panicinfo.PANIC_RECOVERED_TAG, "[JANITOR_WORKER] panic recovered"),
								"file", info.File,
								"line", info.Line,
								"error", info.Message,
							)
						}
					}()
					if err := j.recoverStuckEventConsumers(ctx); err != nil {
						logger.ErrorWithCtx(ctx, "[JANITOR_WORKER] Error recovering stuck event consumers", "error", err)
					}
				}()
			}
		}
	}()
}

func (j *JanitorWorker) recoverStuckEventConsumers(ctx context.Context) error {
	cutoff := time.Now().Add(-j.threshold)
	affected, err := j.eventConsumerRepo.UpdateStuckEventConsumers(ctx, cutoff)
	if err != nil {
		return err
	}
	if affected > 0 {
		logger.InfoWithCtx(ctx, "[JANITOR_WORKER] Recovered stuck event consumers", "affected", affected)
	}
	return nil
}
