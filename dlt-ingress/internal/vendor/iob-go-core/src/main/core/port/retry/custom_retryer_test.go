package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var defaultError = fmt.Errorf("test error")
var retryer = CustomRetryer{}
var attempts = 0
var successOnAttempt = 0
var errorToReturn = defaultError

var functionToRetry = func(ctx context.Context) (any, error) {
	attempts++ // Use the number of attempts for the return value of the function for testing purposes
	if successOnAttempt == attempts {
		return attempts, nil
	}
	return nil, errorToReturn
}

func TestCoreRetryer_Success_On_First_Attempt(t *testing.T) {
	attempts = 0 // Reset number of attempts
	successOnAttempt = 1
	errorToReturn = defaultError
	opts := NewOptions(WithDelay(1*time.Second), WithMaxAttempts(3), WithMultiplier(1.5))

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, successOnAttempt)
	assert.Equal(t, returnedAttempts, successOnAttempt)
	assert.Nil(t, err)
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Success_After_All_Attempts(t *testing.T) {
	attempts = 0 // Reset number of attempts
	successOnAttempt = 3
	errorToReturn = defaultError
	opts := NewOptions(WithDelay(1*time.Second), WithMaxAttempts(3), WithMultiplier(1.5))

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, successOnAttempt)
	assert.Equal(t, returnedAttempts, successOnAttempt)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, elapsed.Seconds(), 2.5)
}

func TestCoreRetryer_Success_After_Some_Attempts(t *testing.T) {
	attempts = 0 // Reset number of attempts
	successOnAttempt = 2
	errorToReturn = defaultError
	opts := NewOptions(WithDelay(1*time.Second), WithMaxAttempts(3), WithMultiplier(2))

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, successOnAttempt)
	assert.Equal(t, returnedAttempts, successOnAttempt)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Error_After_All_Attempts(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(WithDelay(1*time.Second), WithMaxAttempts(3), WithMultiplier(1.5))

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 3)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.GreaterOrEqual(t, elapsed.Seconds(), 1.5)
}

func TestCoreRetryer_Exclude_Error(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(2),
		WithExclude([]error{errorToReturn}),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 1)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Exclude_Error_No_Exclusion(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(2),
		WithExclude([]error{fmt.Errorf("Another error")}),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 3)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.GreaterOrEqual(t, elapsed.Seconds(), 3.0)
}

func TestCoreRetryer_Exclude_Error_By_Logic(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(2),
		WithExcludeLogic(func(err error) bool {
			return err.Error() == defaultError.Error()
		}),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 1)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Exclude_Error_By_Logic_No_Exclusion(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(3),
		WithExcludeLogic(func(err error) bool {
			return err.Error() == "Another error"
		}),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 3)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.GreaterOrEqual(t, elapsed.Seconds(), 4.0)
}

func TestCoreRetryer_Invalid_Max_Attempts(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithMaxAttempts(-1),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 0)
	assert.Nil(t, returnedAttempts)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("invalid maxAttempts %d. Should be 1 or greater", opts.maxAttempts))
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Invalid_Delay(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithDelay(-1),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 0)
	assert.Nil(t, returnedAttempts)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("invalid delay %d. Should be 1 second or greater", opts.delay))
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Invalid_Max_Delay(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithMaxDelay(-1),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 0)
	assert.Nil(t, returnedAttempts)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("invalid maxDelay %d. Should be 1 second or greater", opts.maxDelay))
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Invalid_Multiplier(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithMultiplier(0.5),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 0)
	assert.Nil(t, returnedAttempts)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("invalid multiplier %v. Should be 1.0 or greater", opts.multiplier))
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Cancel_Context(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithMaxAttempts(3),
		WithDelay(10*time.Second),
	)
	ctx, cancel := context.WithCancel(context.Background())

	start := time.Now()
	var elapsed time.Duration
	var returnedAttempts any
	var err error
	go func() {
		returnedAttempts, err = retryer.Execute(ctx, opts, functionToRetry)
		elapsed = time.Since(start)
	}()
	// Cancel context after 2 seconds, retry should be waiting until 2 attempt
	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)

	assert.Equal(t, attempts, 1)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err.Error(), context.Canceled.Error())
	assert.GreaterOrEqual(t, elapsed.Seconds(), 2.0)
}

func TestCoreRetryer_Context_Timeout(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	opts := NewOptions(
		WithMaxAttempts(3),
		WithDelay(10*time.Second),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	returnedAttempts, err := retryer.Execute(ctx, opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 1)
	assert.Nil(t, returnedAttempts)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), context.DeadlineExceeded.Error())
	assert.GreaterOrEqual(t, elapsed.Seconds(), 2.0)
}

func TestCoreRetryer_Exclude_Wrapped_Sentinel_Error(t *testing.T) {
	attempts = 0
	successOnAttempt = 0
	dummySentinelError := errors.New("dummy error")
	errorToReturn = fmt.Errorf("wrapped: %w", dummySentinelError)
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(2),
		WithExclude([]error{dummySentinelError}),
	)

	start := time.Now()
	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)
	elapsed := time.Since(start)

	assert.Equal(t, attempts, 1)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestCoreRetryer_Execute_Failed_Attempt_Func(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	failedAttemptFuncExecMap := map[int]error{}
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(1.5),
		WithFailAttemptFunc(func(ctx context.Context, attempt int, err error) error {
			failedAttemptFuncExecMap[attempt] = err
			return nil
		}),
	)

	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)

	assert.Equal(t, attempts, 3)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.Equal(t, attempts, len(failedAttemptFuncExecMap))
	assert.Equal(t, errorToReturn, failedAttemptFuncExecMap[1])
	assert.Equal(t, errorToReturn, failedAttemptFuncExecMap[2])
	assert.Equal(t, errorToReturn, failedAttemptFuncExecMap[3])
}

func TestCoreRetryer_Execute_Failed_Attempt_Func_With_Error_Exclusion(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 0 // To force to fail in all attempts
	errorToReturn = defaultError
	failedAttemptFuncExecMap := map[int]error{}
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(1.5),
		WithExclude([]error{defaultError}),
		WithFailAttemptFunc(func(ctx context.Context, attempt int, err error) error {
			failedAttemptFuncExecMap[attempt] = err
			return nil
		}),
	)

	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)

	assert.Equal(t, attempts, 1)
	assert.Nil(t, returnedAttempts)
	assert.Equal(t, err, errorToReturn)
	assert.Equal(t, attempts, len(failedAttemptFuncExecMap))
	assert.Equal(t, errorToReturn, failedAttemptFuncExecMap[1])
}

func TestCoreRetryer_Error_On_Failed_Attempt_Func(t *testing.T) {
	attempts = 0         // Reset number of attempts
	successOnAttempt = 2 // To force to fail in all attempts
	errorToReturn = defaultError
	failedAttemptFuncError := fmt.Errorf("failed attempt function error")
	opts := NewOptions(
		WithDelay(1*time.Second),
		WithMaxAttempts(3),
		WithMultiplier(1.5),
		WithFailAttemptFunc(func(ctx context.Context, attempt int, err error) error {
			return failedAttemptFuncError
		}),
	)

	returnedAttempts, err := retryer.Execute(context.Background(), opts, functionToRetry)

	assert.Equal(t, attempts, successOnAttempt)
	assert.Equal(t, returnedAttempts, successOnAttempt)
	assert.Nil(t, err)
}
