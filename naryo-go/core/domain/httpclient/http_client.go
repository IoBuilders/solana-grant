package httpclient

import (
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

type HttpClient struct {
	MaxIdleConnections       int
	KeepAliveDuration        time.Duration
	ConnectTimeout           time.Duration
	ReadTimeout              time.Duration
	WriteTimeout             time.Duration
	CallTimeout              time.Duration
	PingInterval             time.Duration
	RetryOnConnectionFailure bool
}

func NewHttpClient(maxIdleConnections int,
	keepAliveDuration time.Duration,
	connectTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	callTimeout time.Duration,
	pingInterval time.Duration,
	retryOnConnectionFailure bool) *HttpClient {
	return &HttpClient{
		MaxIdleConnections:       maxIdleConnections,
		KeepAliveDuration:        keepAliveDuration,
		ConnectTimeout:           connectTimeout,
		ReadTimeout:              readTimeout,
		WriteTimeout:             writeTimeout,
		CallTimeout:              callTimeout,
		PingInterval:             pingInterval,
		RetryOnConnectionFailure: retryOnConnectionFailure,
	}
}

func (hc *HttpClient) validate() error {
	if hc.MaxIdleConnections <= 0 {
		return domainerrors.NewInvalidFieldError("MaxIdleConnections", "HttpClient", "must be > 0")
	}
	if hc.KeepAliveDuration <= 0 {
		return domainerrors.NewInvalidFieldError("KeepAliveDuration", "HttpClient", "must be > 0")
	}
	if hc.ConnectTimeout <= 0 {
		return domainerrors.NewInvalidFieldError("ConnectTimeout", "HttpClient", "must be > 0")
	}
	if hc.ReadTimeout <= 0 {
		return domainerrors.NewInvalidFieldError("ReadTimeout", "HttpClient", "must be > 0")
	}
	if hc.WriteTimeout <= 0 {
		return domainerrors.NewInvalidFieldError("WriteTimeout", "HttpClient", "must be > 0")
	}
	if hc.CallTimeout <= 0 {
		return domainerrors.NewInvalidFieldError("CallTimeout", "HttpClient", "must be > 0")
	}
	if hc.PingInterval <= 0 {
		return domainerrors.NewInvalidFieldError("PingInterval", "HttpClient", "must be > 0")
	}
	return nil
}
