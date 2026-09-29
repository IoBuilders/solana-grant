package dltingressconfig

import (
	"context"
	"dlt-ingress/src/main/dltingress/config/metrics"
	"dlt-ingress/src/main/dltingress/domain/queuelock"
	"fmt"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/core/shared"
	"dlt-ingress/src/main/core/utils"
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/port/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/port/command/adapters"
	"dlt-ingress/src/main/dltingress/port/contractcallbuilder"
	"dlt-ingress/src/main/dltingress/port/contractcaller"
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/port/custody"
	"dlt-ingress/src/main/dltingress/port/evm"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	"dlt-ingress/src/main/dltingress/port/query/adapters"
	"dlt-ingress/src/main/dltingress/port/repository"
	"dlt-ingress/src/main/dltingress/port/svm"
	"dlt-ingress/src/main/dltingress/port/transactiongasestimator"
	"dlt-ingress/src/main/dltingress/port/transanctionsender"
	"os"

	querygetfailedevents "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/app/query/getfailedevents"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
	"gorm.io/gorm"
)

// TODO: Add a interface common to all bounded contexts
const BoundedContextName = "dltingress"

type DltIngress struct {
	DltIngressQueryBus                query.Bus
	CommandBus                        command.Bus
	EventBus                          event.Bus
	ListenerRegistry                  *event.ListenerRegistry
	CrossListenerRegistry             *event.ListenerRegistry
	DB                                *gorm.DB
	Repositories                      *repository.DltIngressRepositories
	JanitorWorker                     *event.JanitorWorker
	CrossJanitorWorker                *event.JanitorWorker
	BoundedBlockingQueueJanitorWorker *boundedblockingqueue.JanitorWorker
	EventStoreWorker                  *event.StoreWorker
	CrossEventStoreWorker             *event.StoreWorker
	Relay                             event.Relay
	CrossRelay                        event.Relay
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

func WithRepositories(repositories *repository.DltIngressRepositories) DltIngressOption {
	return func(i *DltIngress) { i.Repositories = repositories }
}

func WithCustodyProvider(custodyProvider custody.Port) DltIngressOption {
	return func(i *DltIngress) { i.CustodyProvider = custodyProvider }
}

// Setup initializes the DltIngress components
func Setup(ctx context.Context, core *shared.CoreCommon, opts ...DltIngressOption) (*DltIngress, func(context.Context) error) {
	var dltIngress DltIngress
	for _, opt := range opts {
		opt(&dltIngress)
	}

	if dltIngress.DB == nil {
		dltIngress.DB, _ = CreateDBConnection(ctx)
	}
	tm := db.NewGormTransactionManager(dltIngress.DB)

	if dltIngress.Repositories == nil {
		dltIngress.Repositories = repository.SetupRepositories(dltIngress.DB, tm)
	}

	if dltIngress.ListenerRegistry == nil {
		dltIngress.ListenerRegistry = event.NewListenerRegistry()
	}

	if dltIngress.CrossListenerRegistry == nil {
		dltIngress.CrossListenerRegistry = event.NewListenerRegistry()
	}

	if dltIngress.EventBus == nil {
		dltIngress.EventBus = event.NewOutboxBus(dltIngress.Repositories.EventStoreRepo, dltIngress.ListenerRegistry, []string{}, core.MetricsRegistry)
	}

	if dltIngress.CustodyProvider == nil {
		provider, err := custody.NewCustodyProvider(custody.Config{
			Provider: config.AppConfig.DltIngress.Custody.Provider,
			Dfns: custody.DfnsConfig{
				BaseUrl:      config.AppConfig.DltIngress.Custody.Dfns.BaseUrl,
				AuthToken:    config.AppConfig.DltIngress.Custody.Dfns.AuthToken,
				CredentialID: config.AppConfig.DltIngress.Custody.Dfns.CredentialID,
				PrivateKey:   config.AppConfig.DltIngress.Custody.Dfns.PrivateKey,
			},
			Kms: custody.KMSConfig{
				Region:      config.AppConfig.DltIngress.Custody.Kms.Region,
				AccessKey:   config.AppConfig.DltIngress.Custody.Kms.AccessKey,
				SecretKey:   config.AppConfig.DltIngress.Custody.Kms.SecretKey,
				Endpoint:    config.AppConfig.DltIngress.Custody.Kms.Endpoint,
				Tags:        mapKMSTags(config.AppConfig.DltIngress.Custody.Kms.Tags),
				AliasPrefix: config.AppConfig.DltIngress.Custody.Kms.AliasPrefix,
			},
		})
		if err != nil {
			logger.ErrorWithCtx(ctx, "failed to create custody provider", "error", err)
			panic(err)
		}
		dltIngress.CustodyProvider = provider
	}

	retryConfig := retry.NewOptions(shared.ToRetryOptionConfigs(config.AppConfig.RetryableListeners.Default)...)
	dltIngress.ContractCallBuilderRegistry = *contractcallbuilder.NewRegistry()
	dltIngress.ContractCallerRegistry = *contractcaller.NewRegistry()
	dltIngress.TransactionBuilderRegistry = *contracttransactionbuilder.NewRegistry()
	dltIngress.TransactionSenderRegistry = *transanctionsender.NewRegistry()
	dltIngress.TransactionGasEstimationRegistry = *transactiongasestimator.NewRegistry()
	dltIngress.EvmClientRegistry = core.EvmClientRegistry
	dltIngress.NonceProvider = nonceprovider.NewDbNonceProvider(dltIngress.Repositories.NonceRepo, dltIngress.EvmClientRegistry)
	dltIngress.SvmClientRegistry = svm.NewClientRegistry()
	dltIngress.BlockhashProvider = svm.NewBlockhashProviderAdapterCache(
		svm.NewBlockhashProviderAdapter(dltIngress.SvmClientRegistry),
		core.Cache)

	dltIngress.BoundedBlockingQueue = boundedblockingqueue.NewBoundedBlockingQueueDbAdapterMetrics(
		boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
			dltIngress.Repositories.QueueLockRepo,
			dltIngress.Repositories.TxQueueSlotRepo,
			core.Retryer,
			config.AppConfig.DltIngress.TxBoundedBlockingQueue.Retry.ToRetryOptions(nil, nil),
			tm,
		),
		core.MetricsRegistry,
	)
	dltIngress.DomainServices = SetupDomainServices(dltIngress.Repositories)
	dltIngress.CommandBus = command.NewCommandBus(tm, nil, core.MetricsRegistry)
	dltIngress.AppServices = SetupAppServices(dltIngress.NonceProvider, dltIngress.BlockhashProvider, dltIngress.TransactionBuilderRegistry, dltIngress.EvmClientRegistry, dltIngress.SvmClientRegistry, dltIngress.TransactionGasEstimationRegistry, dltIngress.CommandBus, dltIngress.Repositories)
	dltIngress.DltIngressQueryBus = query.NewQueryBus()
	config.AppConfig.MustValidateSemaphoreTimeout(BoundedContextName)
	stackCfg := eventStackConfig{BoundedContextName, dltIngress.ListenerRegistry, retryConfig, core, false}
	dltIngress.Relay, dltIngress.EventStoreWorker, dltIngress.JanitorWorker = dltIngress.createEventStack(stackCfg)

	crossStackCfg := eventStackConfig{BoundedContextName, dltIngress.CrossListenerRegistry, retryConfig, core, true}
	dltIngress.CrossRelay, dltIngress.CrossEventStoreWorker, dltIngress.CrossJanitorWorker = dltIngress.createEventStack(crossStackCfg)

	if dltIngress.CrossCommandBus == nil {
		dltIngress.CrossCommandBus = command.NewCrossCommandBus(nil, core.MetricsRegistry)
	}

	if dltIngress.CrossQueryBus == nil {
		dltIngress.CrossQueryBus = query.NewCrossQueryBus()
	}

	if dltIngress.CrossEventBus == nil {
		dltIngress.CrossEventBus = event.NewOutboxCrossBus(dltIngress.Repositories.EventStoreRepo, dltIngress.CrossListenerRegistry, core.MetricsRegistry)
	}

	go dltIngress.startWorkers(ctx)

	dltIngress.JanitorWorker.Start(ctx)
	dltIngress.CrossJanitorWorker.Start(ctx)

	dltIngress.BoundedBlockingQueueJanitorWorker = boundedblockingqueue.NewJanitorWorker(
		config.AppConfig.DltIngress.TxBoundedBlockingQueue.ExpiredCleanInterval,
		dltIngress.Repositories.TxQueueSlotRepo,
	)
	dltIngress.BoundedBlockingQueueJanitorWorker.Start(ctx)

	RegisterEvents(dltIngress.ListenerRegistry)
	RegisterCrossEvents(dltIngress.CrossListenerRegistry)
	RegisterListeners(
		dltIngress.ListenerRegistry,
		core.Retryer,
		config.AppConfig.RetryableListeners.DltIngress,
		dltIngress.CrossEventBus,
		dltIngress.AppServices,
		config.AppConfig.ListenerConfig.DltIngress,
	)
	RegisterCrossListeners(dltIngress.CrossListenerRegistry, core.Retryer, config.AppConfig.RetryableListeners.DltIngress, config.AppConfig.ListenerConfig.DltIngress)

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
		dltIngress.CrossRelay.Wait()
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
func setupCommandBus(dltIngress *DltIngress, core *shared.CoreCommon) {
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
	})
	utils.RegisterQueryHandlers(queryBus, handlers)

	// A different query bus for the query is needed because each bounded context will use its own event consumer repository so the core query bus is not valid. In the future there will not be a core query bus.
	utils.RegisterQueryHandlers(dltIngress.DltIngressQueryBus, []any{querygetfailedevents.NewHandler(dltIngress.Repositories.EventConsumerRepo)})
}

func setupCrossCommandBus(dltIngress *DltIngress) {
	adapterList := adapters.InitCommandAdapters(dltIngress.CommandBus)
	utils.RegisterCrossCommandAdapters(dltIngress.CrossCommandBus, adapterList)
}

func setupCrossQueryBus(dltIngress *DltIngress, queryBus query.Bus) {
	adapterList := queryadapters.InitQueryAdapters(queryBus)
	utils.RegisterCrossQueryAdapters(dltIngress.CrossQueryBus, adapterList)
}

func setupDltRegistries(dltIngress *DltIngress) {
	for _, network := range config.AppConfig.DltIngress.Networks {
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
			case common.EVM:
				txBuilder, err := contracttransactionbuilder.NewEvmContractTransactionBuilder(contractInterface)
				if err != nil {
					return err
				}
				dltIngress.TransactionBuilderRegistry.Register(dlt, contractName, txBuilder)

				callBuilder, err := contractcallbuilder.NewEvmContractCallBuilder(contractInterface)
				if err != nil {
					return err
				}
				dltIngress.ContractCallBuilderRegistry.Register(dlt, contractName, callBuilder)
			case common.SVM:
				contractTransactionBuilder, err := contracttransactionbuilder.NewSvmIdlTransactionBuilder([]byte(contractInterface))
				if err != nil {
					return err
				}
				dltIngress.TransactionBuilderRegistry.Register(dlt, contractName, contractTransactionBuilder)
			default:
				logger.Error(fmt.Sprintf("Unsupported dlt type: %s", dlt))
			}
		}
	}
	return nil
}

func setupQueueLocks(ctx context.Context, dltIngress *DltIngress, tm db.TransactionManager) {
	for _, network := range config.AppConfig.DltIngress.Networks {
		if err := dltIngress.Repositories.QueueLockRepo.CreateIfNotExists(ctx, &queuelock.QueueLock{NetworkId: network.Id}); err != nil {
			panic(fmt.Errorf("failed to configure queue lock: %w", err))
		}
	}
}

type eventStackConfig struct {
	contextName string
	registry    *event.ListenerRegistry
	retryConfig retry.Options
	core        *shared.CoreCommon
	isCross     bool
}

func (s *DltIngress) createEventStack(cfg eventStackConfig) (event.Relay, *event.StoreWorker, *event.JanitorWorker) {
	relay := event.NewMemoryRelay(
		cfg.registry,
		s.Repositories.EventConsumerRepo,
		cfg.core.Tracer,
		cfg.core.Retryer,
		cfg.retryConfig,
		cfg.core.MetricsRegistry,
	).WithSemaphoreWaitTimeout(config.AppConfig.GetSemaphoreWaitTimeout(cfg.contextName))

	worker := event.NewStoreWorker(
		config.AppConfig.GetEventStoreInterval(cfg.contextName),
		config.AppConfig.GetEventStoreBatchSize(cfg.contextName),
		s.Repositories.EventStoreRepo,
		relay,
		cfg.isCross,
	)

	janitor := event.NewJanitorWorker(
		s.Repositories.EventConsumerRepo,
		config.AppConfig.GetJanitorInterval(cfg.contextName),
		config.AppConfig.GetJanitorThreshold(cfg.contextName),
	)

	return relay, worker, janitor
}

func (s *DltIngress) startWorkers(ctx context.Context) {
	workers := map[string]*event.StoreWorker{
		"dlt-ingress":              s.EventStoreWorker,
		"dlt-ingress-shared-cross": s.CrossEventStoreWorker,
	}

	for name, w := range workers {
		go func(workerName string, worker *event.StoreWorker) {
			logger.InfoWithCtx(ctx, "worker starting...", "name", name)
			if err := worker.Start(ctx); err != nil {
				logger.ErrorWithCtx(ctx, "failed to start worker", "name", name, "error", err)
			} else {
				logger.InfoWithCtx(ctx, "worker started", "name", name)
			}
		}(name, w)
	}
}
