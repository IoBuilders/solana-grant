package config

import (
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
)

type HttpClientProperties struct {
	MaxIdleConnections       int           `mapstructure:"maxIdleConnections"`
	KeepAliveDuration        time.Duration `mapstructure:"keepAliveDuration"`
	ConnectTimeout           time.Duration `mapstructure:"connectTimeout"`
	ReadTimeout              time.Duration `mapstructure:"readTimeout"`
	WriteTimeout             time.Duration `mapstructure:"writeTimeout"`
	CallTimeout              time.Duration `mapstructure:"callTimeout"`
	PingInterval             time.Duration `mapstructure:"pingInterval"`
	RetryOnConnectionFailure bool          `mapstructure:"retryOnConnectionFailure"`
}

func (hc *HttpClientProperties) Map() (*httpclient.HttpClient, error) {
	return httpclient.NewHttpClient(hc.MaxIdleConnections, hc.KeepAliveDuration, hc.ConnectTimeout, hc.ReadTimeout, hc.WriteTimeout, hc.CallTimeout, hc.PingInterval, hc.RetryOnConnectionFailure), nil
}

var _ descriptor.HttpClient = (*HttpClientProperties)(nil)
