package broadcasterproducer

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/IBM/sarama"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	coremapping "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/mapping"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
)

type KafkaBroadcasterProducer struct {
	producer sarama.SyncProducer
	mapper   coremapping.Mapper[eventmapper.BlockchainEventSource, []byte]
}

func NewKafkaBroadcasterProducer(kafkaBroadcaster *broadcaster.KafkaBroadcaster, mapper coremapping.Mapper[eventmapper.BlockchainEventSource, []byte]) (*KafkaBroadcasterProducer, error) {
	producer, err := sarama.NewSyncProducer(kafkaBroadcaster.Brokers, nil)
	if err != nil {
		return nil, fmt.Errorf("creating kafka producer: %w", err)
	}

	return &KafkaBroadcasterProducer{producer: producer, mapper: mapper}, nil
}

func (p *KafkaBroadcasterProducer) Produce(ctx context.Context, b corebroadcaster.Broadcaster, _ corebroadcaster.Configuration, e event.Event) error {
	payload, err := p.mapper.Map(eventmapper.BlockchainEventSource{Event: e, Broadcaster: b})
	if err != nil {
		return err
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
			if err := p.produceToTopic(ctx, payload, destination.String()); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("topic %s: %w", destination, err))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return errors.Join(errs...)
}

func (p *KafkaBroadcasterProducer) produceToTopic(ctx context.Context, payload []byte, topic string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
	}

	_, _, err := p.producer.SendMessage(message)
	return err
}

func (p *KafkaBroadcasterProducer) Supports(t corebroadcaster.Type) bool {
	return t == broadcaster.TypeKafka
}

var _ broadcast.Producer = (*KafkaBroadcasterProducer)(nil)
