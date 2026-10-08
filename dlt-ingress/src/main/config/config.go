package config

import (
	_ "embed"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/config"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

var DltIngressConfig *Config

type Config struct {
	Database *coreconfig.DatabaseConfig `mapstructure:"database"`

	RetryableListeners struct {
		Default     *coreconfig.RetryConfig               `mapstructure:"default"`
		DltIngress  *coreconfig.BcRetryableListenerConfig `mapstructure:"dltingress"`
		CrossShared *coreconfig.BcRetryableListenerConfig `mapstructure:"crossshared"`
	} `mapstructure:"retryableListeners"`

	ListenerConfig struct {
		Defaults    coreconfig.ListenerConfigDefaults `mapstructure:"defaults"`
		DltIngress  *coreconfig.BcListenerConfig      `mapstructure:"dltingress"`
		CrossShared *coreconfig.BcListenerConfig      `mapstructure:"crossshared"`
	} `mapstructure:"listenerConfig"`

	Janitor struct {
		Default    *coreconfig.JanitorConfig `mapstructure:"default"`
		DltIngress *coreconfig.JanitorConfig `mapstructure:"dltingress"`
	} `mapstructure:"janitor"`

	EventStore struct {
		Default    *coreconfig.EventStoreConfig `mapstructure:"default"`
		DltIngress *coreconfig.EventStoreConfig `mapstructure:"dltingress"`
	} `mapstructure:"eventStore"`

	// Cache carries the TTL for each named cache.Port entry this module reads from or writes to (e.g.
	// "blockhash"). It only configures the cache.Port that gets injected as CoreDependencies.Cache — it
	// does not construct one — so whoever builds that cache.Port for a given deployment must read these
	// TTLs into it, the way LoadConfig does for the bundled examples and integration tests.
	Cache struct {
		Caches map[string]time.Duration `mapstructure:"caches"`
	} `mapstructure:"cache"`

	DltIngress `mapstructure:",squash"`
}

type DltIngress struct {
	AccountRefunding       bool             `mapstructure:"accountRefunding"`
	Networks               []*NetworkConfig `mapstructure:"networks"`
	Custody                CustodyConfig    `mapstructure:"custody"`
	TxBoundedBlockingQueue struct {
		ExpirationTime       time.Duration          `mapstructure:"expirationTime"`
		ExpiredCleanInterval time.Duration          `mapstructure:"expiredCleanInterval"`
		Retry                coreconfig.RetryConfig `mapstructure:"retry"`
	} `mapstructure:"txBoundedBlockingQueue"`
	// LockTimeoutRetryableListeners configures the retry policy shared by every listener registered via
	// ListenerRegistrar.RegisterLockTimeoutRetryable — any listener whose DB write can hit a lock timeout
	// (SQLSTATE 55P03), not just order/DLT-nonce listeners.
	LockTimeoutRetryableListeners struct {
		Retry coreconfig.RetryConfig `mapstructure:"retry"`
	} `mapstructure:"lockTimeoutRetryableListeners"`
	// RateLimitRetryableListeners configures the retry policy for listeners whose node RPC call is
	// rejected by the network rate limiter (ratelimit.ExceededError), independent of the lock-timeout one.
	RateLimitRetryableListeners struct {
		Retry coreconfig.RetryConfig `mapstructure:"retry"`
	} `mapstructure:"rateLimitRetryableListeners"`
}

type NetworkConfig struct {
	Id                   string        `mapstructure:"id"`
	Url                  string        `mapstructure:"url"`
	Dlt                  string        `mapstructure:"dlt"`
	ChainId              amount.Amount `mapstructure:"chainId"`
	TransactionType      uint          `mapstructure:"transactionType"`
	GasLimit             amount.Amount `mapstructure:"gasLimit"`
	GasLimitMultiplier   amount.Amount `mapstructure:"gasLimitMultiplier"`
	GasPrice             amount.Amount `mapstructure:"gasPrice"`
	MaxPriorityFeePerGas amount.Amount `mapstructure:"maxPriorityFeePerGas"`
	FeeMultiplier        amount.Amount `mapstructure:"feeMultiplier"`
	MaxTxPoolSize        int           `mapstructure:"maxTxPoolSize"`
	MaxCuPrice           amount.Amount `mapstructure:"maxCuPrice"`
	// MaxCuLimit is the SVM compute-unit limit ceiling, analogous to GasLimit for EVM. It must stay
	// within Solana's hard per-transaction cap (1,400,000 compute units); GasLimit is not reused here
	// because it is sized for EVM gas, which is a different unit on a different scale.
	MaxCuLimit amount.Amount `mapstructure:"maxCuLimit"`
	// BlockTime is the expected time for this network to mine a new block. It is used to make sure that,
	// for a given wallet, the next nonce is not handed out before the previous one had time to be mined.
	BlockTime time.Duration `mapstructure:"blockTime"`
	// RateLimit is the node's RPC quota, shared by every client talking to this network. Nil means unlimited.
	RateLimit *RateLimitConfig `mapstructure:"rateLimit"`
	Contracts struct {
		Factory                           string `mapstructure:"factory"`
		BusinessLogicResolver             string `mapstructure:"businessLogicResolver"`
		PrimaryMarketFactory              string `mapstructure:"primaryMarketFactory"`
		SettlementHubFactory              string `mapstructure:"settlementHubFactory"`
		ExternalListFactory               string `mapstructure:"externalListFactory"`
		ExternalListConfigurationId       string `mapstructure:"externalListConfigurationId"`
		ExternalListConfigurationVersion  uint64 `mapstructure:"externalListConfigurationVersion"`
		PrimaryMarketConfigurationId      string `mapstructure:"primaryMarketConfigurationId"`
		PrimaryMarketConfigurationVersion uint64 `mapstructure:"primaryMarketConfigurationVersion"`
		SettlementHubConfigurationId      string `mapstructure:"settlementHubConfigurationId"`
		SettlementHubConfigurationVersion uint64 `mapstructure:"settlementHubConfigurationVersion"`
	} `mapstructure:"contracts"`
}

type RetryAfterHeaderFormat string

const (
	RetryAfterHeaderFormatSeconds       RetryAfterHeaderFormat = "seconds"
	RetryAfterHeaderFormatUnixTimestamp RetryAfterHeaderFormat = "unix-timestamp"
)

type RateLimitConfig struct {
	RetryAfterHeader       string                  `mapstructure:"retryAfterHeader"`
	RetryAfterHeaderFormat RetryAfterHeaderFormat  `mapstructure:"retryAfterHeaderFormat"`
	RetryAfterMultiplier   float64                 `mapstructure:"retryAfterMultiplier"`
	Buckets                []RateLimitBucketConfig `mapstructure:"buckets"`
}

type RateLimitBucketConfig struct {
	Name string `mapstructure:"name"`
	// RequestsPerMinute is the bucket's budget. 0 means unlimited.
	RequestsPerMinute int      `mapstructure:"requestsPerMinute"`
	Methods           []string `mapstructure:"methods"`
	MatchAll          bool     `mapstructure:"matchAll"`
}

type KMSTagConfig struct {
	Key   string `mapstructure:"key" json:"key"`
	Value string `mapstructure:"value" json:"value"`
}

type CustodyConfig struct {
	Provider string `mapstructure:"provider"`
	// RequestTimeout bounds each call to the custody provider (a Dfns HTTP request or a whole KMS call, SDK
	// retries included). Zero means the provider package's default.
	RequestTimeout      time.Duration    `mapstructure:"requestTimeout"`
	SignPollingTimeout  time.Duration    `mapstructure:"signPollingTimeout"`
	SignPollingInterval time.Duration    `mapstructure:"signPollingInterval"`
	RateLimit           *RateLimitConfig `mapstructure:"rateLimit"`
	Dfns                struct {
		BaseUrl      string `mapstructure:"baseUrl"`
		AuthToken    string `mapstructure:"authToken"`
		CredentialID string `mapstructure:"credentialId"`
		PrivateKey   string `mapstructure:"privateKey"`
	} `mapstructure:"dfns"`
	Kms struct {
		Region      string         `mapstructure:"region"`
		AccessKey   string         `mapstructure:"accessKey"`
		SecretKey   string         `mapstructure:"secretKey"`
		Endpoint    string         `mapstructure:"endpoint"`
		Tags        []KMSTagConfig `mapstructure:"tags"`
		AliasPrefix string         `mapstructure:"aliasPrefix"`
	} `mapstructure:"kms"`
}

// MustValidateSignPollingTimeoutInDltIngress panics at startup if the dlt ingress sign polling timeout
// is higher than dlt ingress database transaction timeout. This prevents order commands to fail when
// transaction timeout is reached when it is in the async synging polling.
func (c *Config) MustValidateSignPollingTimeoutInDltIngress() {
	if DltIngressConfig.DltIngress.Custody.SignPollingTimeout > 0 && DltIngressConfig.DltIngress.Custody.SignPollingTimeout >= DltIngressConfig.Database.IdleInTransactionSessionTimeout {
		panic(fmt.Sprintf("[dltingress] dltingress.custody.signPollingTimeout (%s) must be < database.dltingress.idleInTransactionSessionTimeout (%s)", DltIngressConfig.DltIngress.Custody.SignPollingTimeout, DltIngressConfig.Database.IdleInTransactionSessionTimeout))
	}
}

// MustValidateCustodyRequestTimeoutInDltIngress panics at startup if a configured custody request timeout
// is not lower than the dlt ingress database transaction timeout: the custody call runs inside that
// transaction, so a longer timeout would let the transaction die before the call gives up.
func (c *Config) MustValidateCustodyRequestTimeoutInDltIngress() {
	timeout := c.DltIngress.Custody.RequestTimeout
	if timeout > 0 && timeout >= c.Database.IdleInTransactionSessionTimeout {
		panic(fmt.Sprintf("[dltingress] dltingress.custody.requestTimeout (%s) must be < database.dltingress.idleInTransactionSessionTimeout (%s)", timeout, c.Database.IdleInTransactionSessionTimeout))
	}
}

func (c *DltIngress) GetDltIngressNetwork(networkId string) (*NetworkConfig, error) {
	for _, network := range c.Networks {
		if network.Id == networkId {
			return network, nil
		}
	}
	return nil, domainerrors.NewEntityNotFoundDomainError("Network", networkId)
}

func (c *DltIngress) GetDltIngressNetworksByDlt(dlt string) []*NetworkConfig {
	var networks []*NetworkConfig
	for _, network := range c.Networks {
		if network.Dlt == dlt {
			networks = append(networks, network)
		}
	}
	return networks
}

func (c *DltIngress) GetNetworkOrDefault(networkId *string) (*NetworkConfig, error) {
	if networkId != nil {
		return c.GetDltIngressNetwork(*networkId)
	}
	return c.Networks[0], nil
}

func (n *NetworkConfig) HasGasCap() bool {
	return !n.GasLimit.IsZero()
}

func (n *NetworkConfig) HasCuLimitCap() bool {
	return !n.MaxCuLimit.IsZero()
}

func (n *NetworkConfig) CheckDltAccount(dltAccountId string, dlt string) error {
	if n.Dlt != dlt {
		return coreerror.NewConflictDomainError(
			"DLT_ACCOUNT_ID_NOT_VALID_FOR_NETWORK",
			fmt.Sprintf("DLT Account Id %s of type %s is not valid for network %s. Type should be %s", dltAccountId, dlt, n.Id, n.Dlt),
		)
	}
	return nil
}

func (c *RateLimitConfig) ToOptionConfigs() []ratelimit.OptionConfig {
	optConfs := make([]ratelimit.OptionConfig, 0, len(c.Buckets))
	for _, b := range c.Buckets {
		matcher := ratelimit.Methods(b.Methods...)
		if b.MatchAll {
			matcher = ratelimit.MatchAll()
		}
		optConfs = append(optConfs, ratelimit.WithBucket(b.Name, b.RequestsPerMinute, matcher))
	}
	return optConfs
}

func (c *RateLimitConfig) Validate() error {
	if c.RetryAfterHeader != "" &&
		c.RetryAfterHeaderFormat != RetryAfterHeaderFormatSeconds &&
		c.RetryAfterHeaderFormat != RetryAfterHeaderFormatUnixTimestamp {
		return fmt.Errorf("retryAfterHeaderFormat must be %q or %q, got %q",
			RetryAfterHeaderFormatSeconds, RetryAfterHeaderFormatUnixTimestamp, c.RetryAfterHeaderFormat)
	}
	if c.RetryAfterMultiplier != 0 && c.RetryAfterMultiplier < 1 {
		return fmt.Errorf("retryAfterMultiplier must be >= 1, got %v", c.RetryAfterMultiplier)
	}

	names := make(map[string]struct{}, len(c.Buckets))
	methods := make(map[string]string)
	for i, b := range c.Buckets {
		if b.Name == "" {
			return fmt.Errorf("bucket at position %d has no name", i)
		}
		if _, exists := names[b.Name]; exists {
			return fmt.Errorf("bucket %s is duplicated", b.Name)
		}
		names[b.Name] = struct{}{}

		if b.RequestsPerMinute < 0 {
			return fmt.Errorf("bucket %s: requestsPerMinute must be >= 0 (0 = unlimited), got %d", b.Name, b.RequestsPerMinute)
		}
		if b.MatchAll == (len(b.Methods) > 0) {
			return fmt.Errorf("bucket %s must declare either methods or matchAll, not both nor neither", b.Name)
		}
		if b.MatchAll && i != len(c.Buckets)-1 {
			return fmt.Errorf("bucket %s matches all methods, so it must be the last one", b.Name)
		}
		for _, method := range b.Methods {
			if owner, exists := methods[method]; exists {
				return fmt.Errorf("method %s is in buckets %s and %s", method, owner, b.Name)
			}
			methods[method] = b.Name
		}
	}
	return nil
}

func (c *DltIngress) GetDefaultDlt() string {
	if len(c.Networks) == 0 {
		return ""
	}
	return c.Networks[0].Dlt
}
