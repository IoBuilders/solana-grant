package httpbroadcasterproducer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-http/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	coremapping "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/mapping"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"
)

// HttpProducer delivers events to a Broadcaster's target destinations via
// HTTP requests, using the endpoint and headers described by the
// broadcaster.HttpConfiguration passed to Produce. One HTTP request is
// sent per destination, concurrently, retried on transient failures.
type HttpProducer struct {
	client *http.Client
	mapper coremapping.Mapper[eventmapper.BlockchainEventSource, []byte]
}

// NewHttpProducer creates a Producer that delivers events over HTTP.
// If client is nil, a default client with a defaultHTTPTimeoutSeconds
// timeout is used.
func NewHttpProducer(cfg *httpclient.HttpClient, mapper coremapping.Mapper[eventmapper.BlockchainEventSource, []byte]) *HttpProducer {
	transport := &http.Transport{
		MaxIdleConns:    cfg.MaxIdleConnections,
		IdleConnTimeout: cfg.KeepAliveDuration,
		DialContext: (&net.Dialer{
			Timeout:   cfg.ConnectTimeout,
			KeepAlive: cfg.KeepAliveDuration,
		}).DialContext,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.ReadTimeout,
	}

	return &HttpProducer{client, mapper}
}

// Produce sends the event to every destination concurrently and waits for
// all of them to finish. Errors from individual destinations are
// collected and returned together, so a failure on one destination
// doesn't stop the rest from being attempted.
func (p *HttpProducer) Produce(ctx context.Context, b corebroadcaster.Broadcaster, configuration corebroadcaster.Configuration, e event.Event) error {
	httpConfig, ok := configuration.(*broadcaster.HTTPConfiguration)
	if !ok {
		return fmt.Errorf("unexpected configuration type %T for HTTP producer", configuration)
	}

	payload, err := p.mapper.Map(eventmapper.BlockchainEventSource{Event: e, Broadcaster: b})
	if err != nil {
		return fmt.Errorf("mapping event: %w", err)
	}

	destinations := b.Target.Destinations()

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	wg.Add(len(destinations))
	for _, destination := range destinations {
		go func() {
			defer wg.Done()
			if err := p.handleSingleDestination(ctx, payload, httpConfig, destination); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("destination %v: %w", destination, err))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return errors.Join(errs...)
}

// handleSingleDestination sends the given payload to a single destination,
// retrying transient failures with exponential backoff.
func (p *HttpProducer) handleSingleDestination(ctx context.Context, payload []byte, httpConfig *broadcaster.HTTPConfiguration, destination target.Destination) error {
	url := httpConfig.Connection.Endpoint.URL() + common.CleanPath(destination.String())

	var lastErr error
	for attempt := 0; attempt <= httpConfig.Connection.Retry.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := p.waitBeforeRetry(ctx, httpConfig.Connection.Retry, attempt, url); err != nil {
				return err
			}
		}

		retryable, err := p.sendOnce(ctx, url, httpConfig, payload)
		if err == nil {
			return nil
		}

		lastErr = err
		if !retryable {
			break
		}
	}

	return lastErr
}

func (p *HttpProducer) waitBeforeRetry(ctx context.Context, retry *common.RetryConfiguration, attempt int, url string) error {
	raw := float64(retry.InitialDelay) * math.Pow(retry.Multiplier, float64(attempt-1))
	delay := time.Duration(math.Min(raw, float64(retry.MaxDelay)))
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while retrying HTTP broadcaster: url=%s: %w", url, ctx.Err())
	}
}

// sendOnce performs a single HTTP attempt. It returns whether the failure
// (if any) is worth retrying: network errors and 5xx responses are
// retryable, 4xx responses are not.
func (p *HttpProducer) sendOnce(ctx context.Context, url string, httpConfig *broadcaster.HTTPConfiguration, payload []byte) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range httpConfig.Connection.Endpoint.Headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return true, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return false, nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	statusErr := fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))

	return resp.StatusCode >= 500, statusErr
}

func (p *HttpProducer) Supports(t corebroadcaster.Type) bool {
	return t == broadcaster.TypeHTTP
}

var _ broadcast.Producer = (*HttpProducer)(nil)
