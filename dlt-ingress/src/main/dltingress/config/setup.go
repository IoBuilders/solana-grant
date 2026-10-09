package dltingressconfig

import (
	"context"
	"fmt"

	"dlt-ingress/src/main/dltingress/config/metrics"
	"dlt-ingress/src/main/dltingress/internal/domain/queuelock"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"

	"os"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/core/shared"
	"dlt-ingress/src/main/core/utils"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/infra/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/internal/infra/contractcallbuilder"
	"dlt-ingress/src/main/dltingress/internal/infra/contractcaller"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	"dlt-ingress/src/main/dltingress/internal/infra/repository"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator"
	"dlt-ingress/src/main/dltingress/internal/infra/transanctionsender"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/app/query/getfailedevents"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/cache"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/config"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const BoundedContextName = "dltingress"

type DltIngress struct {
	DltIngressQueryBus                query.Bus
	CommandBus                        command.Bus
	EventBus                          event.Bus
	ListenerRegistry                  *event.ListenerRegistry
	DB                                *gorm.DB
	Repositories                      *DltIngressRepositories
	JanitorWorker                     *event.JanitorWorker
	BoundedBlockingQueueJanitorWorker *boundedblockingqueue.JanitorWorker
	EventStoreWorker                  *event.StoreWorker
	Relay                             event.Relay
	CrossCommandBus                   command.CrossBus
	CrossQueryBus                     query.CrossBus
	CrossEventBus                     event.CrossBus
	CustodyProvider                   custody.Port
	BoundedBlockingQueue              boundedblockingqueue.Port
	NonceProvider                     nonceprovider.Port
	BlockhashProvider                 svm.BlockhashProvider
	DomainServices                    *DomainServices
	AppServices                       *AppServices
	ContractCallBuilderRegistry       contractcallbuilder.Registry
	ContractCallerRegistry            contractcaller.Registry
	TransactionBuilderRegistry        contracttransactionbuilder.Registry
	TransactionSenderRegistry         transanctionsender.Registry
	TransactionGasEstimationRegistry  transactiongasestimator.Registry
	EvmClientRegistry                 evm.ClientRegistry
	SvmClientRegistry                 svm.ClientRegistry
}

type CoreDependencies struct {
	QueryBus              query.Bus
	CrossQueryBus         query.CrossBus
	CrossCommandBus       command.CrossBus
	CrossEventBus         event.CrossBus
	CrossListenerRegistry *event.ListenerRegistry
	Retryer               retry.Retryer
	Tracer                trace.Tracer
	Cache                 cache.Port
	MetricsRegistry       *metrics.Registry
	HealthRegistry        *health.Registry
}

type DltIngressOption func(ingress *DltIngress)

func WithEventBus(eventBus event.Bus) DltIngressOption {
	return func(i *DltIngress) { i.EventBus = eventBus }
}

func WithListenerRegistry(listenerRegistry *event.ListenerRegistry) DltIngressOption {
	return func(i *DltIngress) { i.ListenerRegistry = listenerRegistry }
}

func WithDB(db *gorm.DB) DltIngressOption {
	return func(i *DltIngress) { i.DB = db }
}

func WithRepositories(repositories *DltIngressRepositories) DltIngressOption {
	return func(i *DltIngress) { i.Repositories = repositories }
}

func WithCustodyProvider(custodyProvider custody.Port) DltIngressOption {
	return func(i *DltIngress) { i.CustodyProvider = custodyProvider }
}

// Setup initializes the DltIngress components
func Setup(ctx context.Context, dltIngressConfig *config.Config, core *CoreDependencies, opts ...DltIngressOption) (*DltIngress, func(context.Context) error) {
	config.DltIngressConfig = dltIngressConfig
	var dltIngress DltIngress
	for _, opt := range opts {
		opt(&dltIngress)
	}

	if dltIngress.DB == nil {
		dltIngress.DB, _ = CreateDBConnection(ctx)
	}
	tm := db.NewGormTransactionManager(dltIngress.DB)

	if dltIngress.Repositories == nil {
		dltIngress.Repositories = SetupRepositories(dltIngress.DB, tm)
	}

	if dltIngress.ListenerRegistry == nil {
		dltIngress.ListenerRegistry = event.NewListenerRegistry()
	}

	if dltIngress.EventBus == nil {
		dltIngress.EventBus = event.NewOutboxBus(dltIngress.Repositories.EventStoreRepo, dltIngress.ListenerRegistry, []string{}, core.MetricsRegistry)
	}

	if dltIngress.CustodyProvider == nil {
		provider, err := custody.NewCustodyProvider(custody.Config{
			Provider: config.DltIngressConfig.DltIngress.Custody.Provider,
			Dfns: custody.DfnsConfig{
				BaseUrl:      config.DltIngressConfig.DltIngress.Custody.Dfns.BaseUrl,
				AuthToken:    config.DltIngressConfig.DltIngress.Custody.Dfns.AuthToken,
				CredentialID: config.DltIngressConfig.DltIngress.Custody.Dfns.CredentialID,
				PrivateKey:   config.DltIngressConfig.DltIngress.Custody.Dfns.PrivateKey,
			},
			Kms: custody.KMSConfig{
				Region:      config.DltIngressConfig.DltIngress.Custody.Kms.Region,
				AccessKey:   config.DltIngressConfig.DltIngress.Custody.Kms.AccessKey,
				SecretKey:   config.DltIngressConfig.DltIngress.Custody.Kms.SecretKey,
				Endpoint:    config.DltIngressConfig.DltIngress.Custody.Kms.Endpoint,
				Tags:        mapKMSTags(config.DltIngressConfig.DltIngress.Custody.Kms.Tags),
				AliasPrefix: config.DltIngressConfig.DltIngress.Custody.Kms.AliasPrefix,
			},
			RateLimit:      config.DltIngressConfig.DltIngress.Custody.RateLimit,
			RequestTimeout: config.DltIngressConfig.DltIngress.Custody.RequestTimeout,
		})
		if err != nil {
			logger.ErrorWithCtx(ctx, "failed to create custody provider", "error", err)
			panic(err)
		}
		dltIngress.CustodyProvider = provider
	}

	retryConfig := retry.NewOptions(shared.ToRetryOptionConfigs(config.DltIngressConfig.RetryableListeners.Default)...)
	dltIngress.ContractCallBuilderRegistry = *contractcallbuilder.NewRegistry()
	dltIngress.ContractCallerRegistry = *contractcaller.NewRegistry()
	dltIngress.TransactionBuilderRegistry = *contracttransactionbuilder.NewRegistry()
	dltIngress.TransactionSenderRegistry = *transanctionsender.NewRegistry()
	dltIngress.TransactionGasEstimationRegistry = *transactiongasestimator.NewRegistry()
	evmNetworksConfigs := config.DltIngressConfig.DltIngress.GetDltIngressNetworksByDlt(string(common.EVM))
	evmClientRegistry, err := evm.NewClientRegistry(evmNetworksConfigs, core.HealthRegistry)
	if err != nil {
		logger.ErrorWithCtx(ctx, "failed to create evm client registry", "error", err)
		panic(err)
	}
	dltIngress.EvmClientRegistry = evmClientRegistry
	dltIngress.NonceProvider = nonceprovider.NewDbNonceProvider(dltIngress.Repositories.NonceRepo, dltIngress.EvmClientRegistry)
	svmNetworksConfigs := config.DltIngressConfig.DltIngress.GetDltIngressNetworksByDlt(string(common.SVM))
	svmClientRegistry, err := svm.NewClientRegistry(svmNetworksConfigs, core.HealthRegistry)
	if err != nil {
		logger.ErrorWithCtx(ctx, "failed to create svm client registry", "error", err)
		panic(err)
	}
	dltIngress.SvmClientRegistry = svmClientRegistry
	dltIngress.BlockhashProvider = svm.NewBlockhashProviderAdapterCache(
		svm.NewBlockhashProviderAdapter(dltIngress.SvmClientRegistry),
		core.Cache)

	dltIngress.BoundedBlockingQueue = boundedblockingqueue.NewBoundedBlockingQueueDbAdapterMetrics(
		boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
			dltIngress.Repositories.QueueLockRepo,
			dltIngress.Repositories.TxQueueSlotRepo,
			core.Retryer,
			config.DltIngressConfig.DltIngress.TxBoundedBlockingQueue.Retry.ToRetryOptions(nil, nil),
			tm,
		),
		core.MetricsRegistry,
	)
	dltIngress.DomainServices = SetupDomainServices(dltIngress.Repositories)
	dltIngress.CommandBus = command.NewCommandBus(tm, nil, core.MetricsRegistry)
	dltIngress.AppServices = SetupAppServices(dltIngress.NonceProvider, dltIngress.BlockhashProvider, dltIngress.TransactionBuilderRegistry, dltIngress.EvmClientRegistry, dltIngress.SvmClientRegistry, dltIngress.TransactionGasEstimationRegistry, dltIngress.CommandBus, dltIngress.Repositories)
	dltIngress.DltIngressQueryBus = query.NewQueryBus()
	coreconfig.MustValidateSemaphoreTimeout(BoundedContextName, config.DltIngressConfig.EventStore.DltIngress, config.DltIngressConfig.EventStore.Default, config.DltIngressConfig.Janitor.DltIngress, config.DltIngressConfig.Janitor.Default)
	config.DltIngressConfig.MustValidateSignPollingTimeoutInDltIngress()
	config.DltIngressConfig.MustValidateCustodyRequestTimeoutInDltIngress()
	dltIngress.Relay = event.NewMemoryRelay(dltIngress.ListenerRegistry, dltIngress.Repositories.EventConsumerRepo, core.Tracer, core.Retryer, retryConfig, core.MetricsRegistry).
		WithSemaphoreWaitTimeout(coreconfig.GetSemaphoreWaitTimeout(config.DltIngressConfig.EventStore.DltIngress, config.DltIngressConfig.EventStore.Default))
	dltIngress.EventStoreWorker = event.NewStoreWorker(
		coreconfig.GetEventStoreInterval(config.DltIngressConfig.EventStore.DltIngress, config.DltIngressConfig.EventStore.Default),
		coreconfig.GetEventStoreBatchSize(config.DltIngressConfig.EventStore.DltIngress, config.DltIngressConfig.EventStore.Default),
		dltIngress.Repositories.EventStoreRepo,
		dltIngress.Relay,
		false,
	)
	go func() {
		err := dltIngress.EventStoreWorker.Start(ctx)
		if err != nil {
			logger.ErrorWithCtx(ctx, "failed to start dltIngress event store worker", "error", err)
		} else {
			logger.InfoWithCtx(ctx, "dltIngress event store worker started")
		}
	}()

	dltIngress.JanitorWorker = event.NewJanitorWorker(
		dltIngress.Repositories.EventConsumerRepo,
		coreconfig.GetJanitorInterval(config.DltIngressConfig.Janitor.DltIngress, config.DltIngressConfig.Janitor.Default),
		coreconfig.GetJanitorThreshold(config.DltIngressConfig.Janitor.DltIngress, config.DltIngressConfig.Janitor.Default),
	)
	dltIngress.JanitorWorker.Start(ctx)

	dltIngress.BoundedBlockingQueueJanitorWorker = boundedblockingqueue.NewJanitorWorker(
		config.DltIngressConfig.DltIngress.TxBoundedBlockingQueue.ExpiredCleanInterval,
		dltIngress.Repositories.TxQueueSlotRepo,
	)
	dltIngress.BoundedBlockingQueueJanitorWorker.Start(ctx)

	dltIngress.CrossCommandBus = core.CrossCommandBus
	dltIngress.CrossQueryBus = core.CrossQueryBus
	dltIngress.CrossEventBus = core.CrossEventBus

	RegisterEvents(dltIngress.ListenerRegistry)
	RegisterCrossEvents(core.CrossListenerRegistry)
	RegisterListeners(
		dltIngress.ListenerRegistry,
		core.Retryer,
		config.DltIngressConfig.RetryableListeners.DltIngress,
		dltIngress.CrossEventBus,
		dltIngress.AppServices,
		config.DltIngressConfig.ListenerConfig.DltIngress,
	)
	RegisterCrossListeners(core.CrossListenerRegistry, core.Retryer, config.DltIngressConfig.RetryableListeners.CrossShared, config.DltIngressConfig.ListenerConfig.CrossShared)

	if err := dltingressmetrics.SetupMetrics(core.MetricsRegistry); err != nil {
		panic(err)
	}
	setupCommandBus(&dltIngress, core)
	setupQueryBus(core.QueryBus, &dltIngress)
	setupCrossCommandBus(&dltIngress)
	setupCrossQueryBus(&dltIngress, core.QueryBus)
	if err := setupDltBuilders(&dltIngress); err != nil {
		panic(err)
	}
	setupDltRegistries(&dltIngress)
	setupQueueLocks(ctx, &dltIngress, tm)

	shutdownFn := func(shutdownCtx context.Context) error {
		logger.InfoWithCtx(shutdownCtx, "Shutting down 'dltIngress' module...")

		if dltIngress.EvmClientRegistry != nil {
			logger.InfoWithCtx(shutdownCtx, "Closing eth clients' connections.")
			dltIngress.EvmClientRegistry.Shutdown()
		}

		if dltIngress.SvmClientRegistry != nil {
			logger.InfoWithCtx(shutdownCtx, "Closing solana clients' connections.")
			if err := dltIngress.SvmClientRegistry.Shutdown(); err != nil {
				return err
			}
		}

		dltIngress.Relay.Wait()
		if dltIngress.DB != nil {
			sqlDB, err := dltIngress.DB.DB()
			if err == nil {
				logger.InfoWithCtx(shutdownCtx, "Closing 'dltIngress' database connection.")
				return sqlDB.Close()
			}
		}
		return nil
	}

	return &dltIngress, shutdownFn
}

func CreateDBConnection(ctx context.Context) (*gorm.DB, error) {
	dltIngressDb, err := repository.NewPostgresDB()
	if err != nil {
		logger.ErrorWithCtx(ctx, "failed to connect to database", "error", err)
		os.Exit(1)
	}
	if err := repository.Migrate(dltIngressDb); err != nil {
		panic(fmt.Sprintf("migration failed: %v", err))
	}
	return dltIngressDb, nil
}

// setupCommandBus initializes the command bus
func setupCommandBus(dltIngress *DltIngress, core *CoreDependencies) {
	handlers := InitCommandHandlers(CommandHandlerDeps{
		Repositories:                     dltIngress.Repositories,
		CustodyProvider:                  dltIngress.CustodyProvider,
		EventBus:                         dltIngress.EventBus,
		DomainServices:                   dltIngress.DomainServices,
		NonceProvider:                    dltIngress.NonceProvider,
		TransactionSenderRegistry:        dltIngress.TransactionSenderRegistry,
		TransactionGasEstimationRegistry: dltIngress.TransactionGasEstimationRegistry,
		BoundedBlockingQueue:             dltIngress.BoundedBlockingQueue,
		EvmClientRegistry:                dltIngress.EvmClientRegistry,
		SvmClientRegistry:                dltIngress.SvmClientRegistry,
		AppServices:                      dltIngress.AppServices,
		MetricsRegistry:                  core.MetricsRegistry,
	})
	utils.RegisterCommandHandlers(dltIngress.CommandBus, handlers)
}

func setupQueryBus(queryBus query.Bus, dltIngress *DltIngress) {
	handlers := SetupQueryHandlers(QueryHandlerDeps{
		Repositories:                dltIngress.Repositories,
		ContractCallBuilderRegistry: dltIngress.ContractCallBuilderRegistry,
		ContractCallerRegistry:      dltIngress.ContractCallerRegistry,
		EvmClientRegistry:           dltIngress.EvmClientRegistry,
		SvmClientRegistry:           dltIngress.SvmClientRegistry,
	})
	utils.RegisterQueryHandlers(queryBus, handlers)

	// A different query bus for the query is needed because each bounded context will use its own event consumer repository so the core query bus is not valid. In the future there will not be a core query bus.
	utils.RegisterQueryHandlers(dltIngress.DltIngressQueryBus, []any{getfailedevents.NewHandler(dltIngress.Repositories.EventConsumerRepo)})
}

func setupCrossCommandBus(dltIngress *DltIngress) {
	adapterList := InitCommandAdapters(dltIngress.CommandBus)
	utils.RegisterCrossCommandAdapters(dltIngress.CrossCommandBus, adapterList)
}

func setupCrossQueryBus(dltIngress *DltIngress, queryBus query.Bus) {
	adapterList := InitQueryAdapters(queryBus)
	utils.RegisterCrossQueryAdapters(dltIngress.CrossQueryBus, adapterList)
}

func setupDltRegistries(dltIngress *DltIngress) {
	for _, network := range config.DltIngressConfig.DltIngress.Networks {
		dlt := common.Dlt(network.Dlt)
		switch dlt {
		case common.EVM:
			evmClientRegistry := dltIngress.EvmClientRegistry
			dltIngress.ContractCallerRegistry.Register(dlt, contractcaller.NewEvmContractCaller(evmClientRegistry))
			dltIngress.TransactionSenderRegistry.Register(dlt, transanctionsender.NewEvmTransactionSender(evmClientRegistry))
			dltIngress.TransactionGasEstimationRegistry.Register(dlt, transactiongasestimator.NewEvmTransactionGasEstimator(evmClientRegistry))
		case common.SVM:
			svmClientRegistry := dltIngress.SvmClientRegistry
			dltIngress.TransactionSenderRegistry.Register(dlt, transanctionsender.NewSvmTransactionSender(svmClientRegistry))
			dltIngress.TransactionGasEstimationRegistry.Register(dlt, transactiongasestimator.NewSvmTransactionGasEstimator(svmClientRegistry))
		default:
			logger.Error(fmt.Sprintf("Unsupported dlt type: %s", dlt))
		}
	}
}

func mapKMSTags(tags []config.KMSTagConfig) []custody.KMSTag {
	result := make([]custody.KMSTag, len(tags))
	for i, t := range tags {
		result[i] = custody.KMSTag{Key: t.Key, Value: t.Value}
	}
	return result
}

func setupDltBuilders(dltIngress *DltIngress) error {
	for dlt, contractDefinitions := range SmartContractDefinitions {
		for contractName, contractInterface := range contractDefinitions {
			switch dlt {
			case string(common.EVM):
				txBuilder, err := contracttransactionbuilder.NewEvmContractTransactionBuilder(contractInterface)
				if err != nil {
					return err
				}
				dltIngress.TransactionBuilderRegistry.Register(common.Dlt(dlt), contractName, txBuilder)

				callBuilder, err := contractcallbuilder.NewEvmContractCallBuilder(contractInterface)
				if err != nil {
					return err
				}
				dltIngress.ContractCallBuilderRegistry.Register(common.Dlt(dlt), contractName, callBuilder)
			case string(common.SVM):
				contractTransactionBuilder, err := contracttransactionbuilder.NewSvmIdlTransactionBuilder([]byte(contractInterface))
				if err != nil {
					return err
				}
				dltIngress.TransactionBuilderRegistry.Register(common.Dlt(dlt), contractName, contractTransactionBuilder)
			default:
				logger.Error(fmt.Sprintf("Unsupported dlt type: %s", dlt))
			}
		}
	}
	return nil
}

func setupQueueLocks(ctx context.Context, dltIngress *DltIngress, tm db.TransactionManager) {
	for _, network := range config.DltIngressConfig.DltIngress.Networks {
		if err := dltIngress.Repositories.QueueLockRepo.CreateIfNotExists(ctx, &queuelock.QueueLock{NetworkId: network.Id}); err != nil {
			panic(fmt.Errorf("failed to configure queue lock: %w", err))
		}
	}
}

func NewInMemoryProvider() custody.Port {
	return custody.NewInMemoryProvider()
}
