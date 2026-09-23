# Configuring `broadcaster-kafka`

`broadcaster-kafka` is an adapter module: it implements `core`'s `broadcast.Producer` port by
publishing events to Kafka topics. Configuring it means two things: pointing it at your Kafka brokers,
and declaring a `KAFKA` broadcaster configuration entry so the core can route broadcasters to it.

## 1. Broker connection

The module reads its own `kafka:` section, under the same root property (typically `naryo`) as
the rest of the application config:

```yaml
naryo:
  kafka:
    brokers:
      - "${KAFKA_BROKER:localhost:29092}"
```

| Key       | Type       | Required | Notes                                                                                                   |
|-----------|------------|----------|---------------------------------------------------------------------------------------------------------|
| `brokers` | `[]string` | yes      | One or more `host:port` addresses. Must be non-empty, and each entry must parse as a valid `host:port`. |

## 2. Broadcaster configuration entry

Separately from the broker connection above, `core`'s generic `broadcasting.configuration`
list needs at least one entry of `type: "KAFKA"` for any `broadcasting.broadcasters` entry
to route to this module:

```yaml
naryo:
  broadcasting:
    configuration:
      - id: "00000000-0000-0000-0000-000000000001"
        type: "KAFKA"
    broadcasters:
      - id: "00000000-0000-0000-0000-000000000002"
        configurationId: "00000000-0000-0000-0000-000000000001"
        target:
          type: "CONTRACT_EVENT"
          destinations:
            - "contract-events"
```

Under `broadcasting.configuration`, only `id` and `type` matter, since this module doesn't have an adapter-specific configuration.

Under `broadcasting.broadcasters`, the Kafka topic a message is published to comes from `target.destinations`.
