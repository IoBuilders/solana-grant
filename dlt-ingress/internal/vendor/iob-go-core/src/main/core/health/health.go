package health

import (
	"context"
	"time"
)

type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

type CheckResult struct {
	Name         string        `json:"-"`
	Status       Status        `json:"status"`
	ResponseTime string 	   `json:"responseTime,omitempty"`
	Error        string        `json:"error,omitempty"`
}

type Response struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

func NewDownCheckResult(err error, startTime time.Time) CheckResult {
	return CheckResult{
		Status:       StatusDown,
		ResponseTime: getResponseTime(startTime),
		Error:        err.Error(),
	}
}

func NewUpCheckResult(startTime time.Time) CheckResult {
	return CheckResult{
		Status:       StatusUp,
		ResponseTime: getResponseTime(startTime),
	}
}

func getResponseTime(startTime time.Time) string {
	// Round to microseconds to clean to always show 2 decimals
	return time.Since(startTime).Round(10 * time.Microsecond).String()
}
