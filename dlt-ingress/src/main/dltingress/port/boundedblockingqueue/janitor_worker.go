package boundedblockingqueue

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/repository/txqueueslot"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/panicinfo"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type JanitorWorker struct {
	interval   time.Duration
	repository txqueueslotrepo.Repository
}

func NewJanitorWorker(interval time.Duration, repository txqueueslotrepo.Repository) *JanitorWorker {
	return &JanitorWorker{
		interval:   interval,
		repository: repository,
	}
}

func (j *JanitorWorker) Start(ctx context.Context) {

	logger.InfoWithCtx(ctx, "[BOUNDED BLOCKING QUEUE JANITOR_WORKER] Transactions bounded blocking queue started with interval", "interval", j.interval)

	ticker := time.NewTicker(j.interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.InfoWithCtx(ctx, "[BOUNDED BLOCKING QUEUE JANITOR_WORKER] Stopped")
				return
			case <-ticker.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							info := panicinfo.NewPanicInfo(r)
							logger.InfoWithCtx(ctx,
								fmt.Sprintf("%s %s", panicinfo.PANIC_RECOVERED_TAG, "[BOUNDED BLOCKING QUEUE JANITOR_WORKER] Recovered from panic"),
								"panic", info.Message,
								"file", info.File,
								"line", info.Line,
								"stack", info.FormatStackTrace(),
							)
						}
					}()
					if err := j.repository.DeleteExpired(ctx); err != nil {
						logger.ErrorWithCtx(ctx, "[BOUNDED BLOCKING QUEUE JANITOR_WORKER] Error deleting expired slots", "error", err)
					}
				}()
			}
		}
	}()
}
