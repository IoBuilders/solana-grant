package broadcasterproducer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rabbitmq/amqp091-go"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	coremapping "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/mapping"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
)

const contentType = "application/json"

type publisher interface {
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error
}

type RabbitMQBroadcasterProducer struct {
	channel publisher
	mapper  coremapping.Mapper[eventmapper.BlockchainEventSource, []byte]
}

func NewRabbitMQBroadcasterProducer(rabbitMQBroadcaster *broadcaster.RabbitMQBroadcaster, mapper coremapping.Mapper[eventmapper.BlockchainEventSource, []byte]) (*RabbitMQBroadcasterProducer, error) {
	connection, err := dial(rabbitMQBroadcaster)
	if err != nil {
		return nil, fmt.Errorf("connecting to rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening rabbitmq channel: %w", err)
	}

	return &RabbitMQBroadcasterProducer{channel: channel, mapper: mapper}, nil
}

func (p *RabbitMQBroadcasterProducer) Produce(ctx context.Context, b corebroadcaster.Broadcaster, configuration corebroadcaster.Configuration, e event.Event) error {
	rabbitMQConfiguration, ok := configuration.(*broadcaster.RabbitMQConfiguration)
	if !ok {
		return fmt.Errorf("unsupported configuration %T for the rabbitmq producer", configuration)
	}

	payload, err := p.mapper.Map(eventmapper.BlockchainEventSource{Event: e, Broadcaster: b})
	if err != nil {
		return err
	}

	var errs []error
	for _, destination := range b.Target.Destinations() {
		routingKey, err := routingKeyFor(e, destination, b.Target.Type())
		if err != nil {
			errs = append(errs, fmt.Errorf("destination %s: %w", destination, err))
			continue
		}
		if err := p.publish(ctx, payload, rabbitMQConfiguration.Exchange, routingKey); err != nil {
			errs = append(errs, fmt.Errorf("exchange %s, routing key %s: %w", rabbitMQConfiguration.Exchange, routingKey, err))
		}
	}

	return errors.Join(errs...)
}

func (p *RabbitMQBroadcasterProducer) publish(ctx context.Context, payload []byte, exchange broadcaster.Exchange, routingKey broadcaster.RoutingKey) error {
	return p.channel.PublishWithContext(ctx, exchange.String(), routingKey.String(), false, false, amqp091.Publishing{
		ContentType: contentType,
		Body:        payload,
	})
}

// routingKeyFor builds "<destination>.<key>", where key identifies e: the slot
// for a block, the signature for a transaction, and "<eventName>.<programId>"
// for a contract event, or just "<programId>" when routed through a FILTER target.
func routingKeyFor(e event.Event, destination target.Destination, targetType target.Type) (broadcaster.RoutingKey, error) {
	var key string
	switch ev := e.(type) {
	case event.SolanaContractEvent:
		if targetType == target.TypeFilter {
			key = ev.ProgramID
		} else {
			key = ev.EventName() + "." + ev.ProgramID
		}
	case event.SolanaBlockEvent:
		key = strconv.FormatUint(ev.Slot, 10)
	case event.SolanaTransactionEvent:
		key = ev.Signature
	default:
		return "", fmt.Errorf("unsupported event type %T", e)
	}
	return broadcaster.NewRoutingKey(destination.String() + "." + key)
}

func (p *RabbitMQBroadcasterProducer) Supports(t corebroadcaster.Type) bool {
	return t == broadcaster.TypeRabbitMQ
}

func dial(b *broadcaster.RabbitMQBroadcaster) (*amqp091.Connection, error) {
	if !b.TLSEnabled {
		return amqp091.Dial(amqpURI(b, "amqp"))
	}

	minVersion, err := tlsMinVersion(b.TLSAlgorithm)
	if err != nil {
		return nil, err
	}

	return amqp091.DialTLS(amqpURI(b, "amqps"), &tls.Config{
		ServerName: b.Host,
		MinVersion: minVersion,
	})
}

func amqpURI(b *broadcaster.RabbitMQBroadcaster, scheme string) string {
	u := url.URL{
		Scheme: scheme,
		User:   url.UserPassword(b.Username, b.Password),
		Host:   b.Host + ":" + strconv.Itoa(b.Port),
	}
	if vhost := strings.TrimPrefix(b.VirtualHost, "/"); vhost != "" {
		u.Path = "/" + vhost
	}
	return u.String()
}

func tlsMinVersion(algorithm broadcaster.TLSAlgorithm) (uint16, error) {
	switch algorithm {
	case broadcaster.TLSAlgorithmTLS12:
		return tls.VersionTLS12, nil
	case broadcaster.TLSAlgorithmTLS13:
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported TLS algorithm %q", algorithm)
	}
}

var _ broadcast.Producer = (*RabbitMQBroadcasterProducer)(nil)
