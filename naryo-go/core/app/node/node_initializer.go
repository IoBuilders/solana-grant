package node

import (
	"context"
	"fmt"
	"reflect"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/dispatch"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/retry"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/routing"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/subscription"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger/blockprocessor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger/slotprocessor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger/transactionprocessor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/solana"
)

type NodeInitializer interface {
	Initialize(ctx context.Context) ([]NodeRunner, error)
}

type DefaultNodeInitializer struct {
	nodeConfigManager                 configurationmanager.NodeConfigurationManager
	broadcasterConfigManager          configurationmanager.BroadcasterConfigurationManager
	broadcasterConfigConfigManager    configurationmanager.BroadcasterConfigurationConfigurationManager
	filterConfigManager               configurationmanager.FilterConfigurationManager
	httpClientConfigManager           configurationmanager.HttpClientConfigurationManager
	storeConfigManager                configurationmanager.StoreConfigurationManager
	programErrorRegistryConfigManager configurationmanager.ProgramErrorRegistryConfigurationManager
	producers                         []broadcast.Producer
	d                                 decoder.Decoder
	retryer                           retry.Retryer
	eventStores                       map[reflect.Type]any
	latestBlockStores                 map[reflect.Type]any
}

func NewDefaultNodeInitializer(
	nodeConfigManager configurationmanager.NodeConfigurationManager,
	broadcasterConfigManager configurationmanager.BroadcasterConfigurationManager,
	broadcasterConfigConfigManager configurationmanager.BroadcasterConfigurationConfigurationManager,
	filterConfigManager configurationmanager.FilterConfigurationManager,
	httpClientConfigManager configurationmanager.HttpClientConfigurationManager,
	storeConfigManager configurationmanager.StoreConfigurationManager,
	programErrorRegistryConfigManager configurationmanager.ProgramErrorRegistryConfigurationManager,
	producers []broadcast.Producer,
	d decoder.Decoder,
	retryer retry.Retryer,
	eventStores map[reflect.Type]any,
	latestBlockStores map[reflect.Type]any,
) *DefaultNodeInitializer {
	return &DefaultNodeInitializer{
		nodeConfigManager:                 nodeConfigManager,
		broadcasterConfigManager:          broadcasterConfigManager,
		broadcasterConfigConfigManager:    broadcasterConfigConfigManager,
		filterConfigManager:               filterConfigManager,
		httpClientConfigManager:           httpClientConfigManager,
		storeConfigManager:                storeConfigManager,
		programErrorRegistryConfigManager: programErrorRegistryConfigManager,
		producers:                         producers,
		d:                                 d,
		retryer:                           retryer,
		eventStores:                       eventStores,
		latestBlockStores:                 latestBlockStores,
	}
}

func (i *DefaultNodeInitializer) Initialize(ctx context.Context) ([]NodeRunner, error) {
	nodes, err := i.nodeConfigManager.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading node configuration: %w", err)
	}
	httpClient, err := i.httpClientConfigManager.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading http client configuration: %w", err)
	}

	nodeRunners := make([]NodeRunner, 0, len(nodes))
	for _, n := range nodes {
		nodeRunner, err := i.initializeNode(ctx, n, httpClient)
		if err != nil {
			logging.ErrorWithCtx(ctx, fmt.Sprintf("failed to initialize node: %s", n.Name), "error", err)
		} else {
			nodeRunners = append(nodeRunners, nodeRunner)
		}
	}

	return nodeRunners, nil
}

func (i *DefaultNodeInitializer) initializeNode(ctx context.Context, n *node.Node, client *httpclient.HttpClient) (NodeRunner, error) {
	dispatcher := dispatch.NewEventDispatcher()
	if err := i.addSharedTriggers(ctx, dispatcher); err != nil {
		return nil, fmt.Errorf("error loading broadcaster routing: %w", err)
	}

	contractHelper := blockprocessor.NewContractEventDispatcherHelper(dispatcher)

	switch n.Type {
	case node.TypeSolana:
		return i.initializeSolanaNode(n, dispatcher, client, contractHelper)
	default:
		return nil, fmt.Errorf("unsupported node type: %s", n.Type)
	}
}

func (i *DefaultNodeInitializer) initializeSolanaNode(n *node.Node, dispatcher dispatch.Dispatcher, client *httpclient.HttpClient, contractHelper *blockprocessor.ContractEventDispatcherHelper) (*DefaultNodeRunner, error) {
	retryConfig := common.DefaultRetryConfiguration()
	switch n.Subscription.MethodConfiguration.Method() {
	case node.BlockSubscriptionMethodPoll:
		blockInteractor, err := solana.NewSolanaRpcBlockInteractor(n.Connection, client)
		if err != nil {
			return nil, err
		}
		startSlotCalculator := subscription.NewSolanaStartSlotCalculator(n, blockInteractor, getStore[event.SolanaBlockEvent, store.LatestBlockStore[event.SolanaBlockEvent]](i.latestBlockStores))
		subscriber := subscription.NewSolanaPollSlotSubscriber(n, blockInteractor, dispatcher, startSlotCalculator, i.retryer, common.DefaultRetryConfiguration())

		i.addSolanaNodeSpecificTriggers(blockInteractor, dispatcher, contractHelper, retryConfig)
		return NewDefaultNodeRunner(n, subscriber, dispatcher), nil
	case node.BlockSubscriptionMethodPubSub:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported subscription method: %s", n.Subscription.MethodConfiguration.Method())
	}
}

// addSharedTriggers wires the EventBroadcasterPermanentTrigger into dispatcher. Both it and its
// Router keep an in-memory configuration snapshot that Process/Match read without touching the
// configuration managers — each requires an explicit Reload to populate that snapshot at all
// (nothing else calls it), so without these two calls no broadcaster or filter would ever match.
func (i *DefaultNodeInitializer) addSharedTriggers(ctx context.Context, dispatcher dispatch.Dispatcher) error {
	router := routing.NewEventRoutingService(i.broadcasterConfigManager, i.filterConfigManager)
	if err := router.Reload(ctx); err != nil {
		return fmt.Errorf("error loading broadcaster/filter routing configuration: %w", err)
	}

	broadcasterTrigger := trigger.NewEventBroadcasterPermanentTrigger(router, i.producers, i.broadcasterConfigConfigManager)
	if err := broadcasterTrigger.Reload(ctx); err != nil {
		return fmt.Errorf("error loading broadcaster configuration: %w", err)
	}

	dispatcher.AddTrigger(broadcasterTrigger)
	return nil
}

func (i *DefaultNodeInitializer) addSolanaNodeSpecificTriggers(
	interactor interactor.BlockInteractor,
	dispatcher dispatch.Dispatcher,
	contractHelper *blockprocessor.ContractEventDispatcherHelper,
	retryConfig *common.RetryConfiguration,
) {
	dispatcher.AddTrigger(
		trigger.NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](
			getStore[event.SolanaBlockEvent, store.BlockEventStore[event.SolanaBlockEvent]](i.eventStores),
			getStore[event.SolanaTransactionEvent, store.TransactionEventStore[event.SolanaTransactionEvent]](i.eventStores),
			getStore[event.SolanaContractEvent, store.ContractEventStore[event.SolanaContractEvent]](i.eventStores),
			i.storeConfigManager,
		),
	)
	dispatcher.AddTrigger(slotprocessor.NewSlotProcessorPermanentTrigger(interactor, dispatcher, retryConfig))
	dispatcher.AddTrigger(blockprocessor.NewSolanaBlockProcessorPermanentTrigger(i.filterConfigManager, i.d, contractHelper))
	dispatcher.AddTrigger(trigger.NewLatestBlockStorePermanentTrigger(
		getStore[event.SolanaBlockEvent, store.LatestBlockStore[event.SolanaBlockEvent]](i.latestBlockStores),
		i.storeConfigManager,
	))
	dispatcher.AddTrigger(transactionprocessor.NewSolanaTransactionProcessorPermanentTrigger(i.filterConfigManager, dispatcher, i.programErrorRegistryConfigManager))
}

func getStore[D any, S any](stores map[reflect.Type]any) S {
	return stores[reflect.TypeFor[D]()].(S)
}

var _ NodeInitializer = (*DefaultNodeInitializer)(nil)
