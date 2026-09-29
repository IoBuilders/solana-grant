package health

import (
	"context"
	"sync"
	"time"
)

type Registry struct {
	mu       sync.RWMutex
	checkers []Checker
}

func NewRegistry() *Registry {
	return &Registry{
		checkers: make([]Checker, 0),
	}
}

func (r *Registry) Register(checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, checker)
}

func (r *Registry) RunAllChecks(ctx context.Context) Response {
	// Read lock to protect checkers slice
	r.mu.RLock()
	checkersCount := len(r.checkers)
	localCheckers := make([]Checker, checkersCount)
	copy(localCheckers, r.checkers)
	r.mu.RUnlock()

	// Execute checkers concurrently
	var wg sync.WaitGroup
	resultsCh := make(chan CheckResult, checkersCount)
	status := StatusUp
	for _, checker := range localCheckers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			res := c.Check(ctx)
			res.Name = c.Name()
			resultsCh <- res
		}(checker)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect check results
	checks := make(map[string]CheckResult, checkersCount)
	for r := range resultsCh {
		checks[r.Name] = r
		if r.Status == StatusDown {
			status = StatusDown
		}
	}

	return Response{
		Status:    status,
		Timestamp: time.Now(),
		Checks:    checks,
	}
}
