# Getting Started with naryo-go

Welcome to naryo-go! This guide will walk you through the first steps to get naryo-go up and
running.

There are two ways to use naryo-go:

1. **Docker (Easiest)**: Run the naryo-go reference server in a Docker container.
2. **Go library (Advanced)**: Import `core` and the adapters you need directly into your own Go
   program. This gives you maximum control but requires manual wiring: naryo-go doesn't have a
   Spring Boot-style autoconfiguration layer the way the Java original does.

## 1. Running naryo-go with Docker

The fastest way to see it running is the Docker-based quickstart:

- [Naryo-go Quickstart: Solana](./tutorials/naryo_quickstart.md)

For a production-shaped setup, `server`'s own Docker Compose stack
([`docker-compose.yml`](../docker-compose.yml) at the repo root) brings up `naryo-server` alongside
Postgres and (for local development) a [surfpool](https://github.com/solana-foundation/surfpool)
Solana RPC node:

```bash
make compose-up
```

Configuration is a YAML file (`CONFIG_PATH` env var) merged over `server`'s embedded default
(`server/cmd/server/resources/application.yml`). See
[`examples/quickstart/application.yaml`](../examples/quickstart/application.yaml) for a complete,
annotated example: nodes, filters, broadcasting, and store configuration all live under a single
`naryo:` root key.

naryo-go doesn't have a dedicated configuration reference doc yet (the Java original documents
every option per node/filter/broadcaster/persistence backend); `core/infrastructure/config`'s
`mapstructure`-tagged Go types are the source of truth in the meantime.

## 2. Using naryo-go as a Go library (Advanced)

`core`, `broadcaster-http`, `broadcaster-kafka` and `persistence-gorm` are ordinary Go modules.
Nothing about them requires running as the `server` binary. `server/cmd/server/main.go` is itself
the reference example of this: strip away its flag parsing and `/health` endpoint, and what's
left is just calling exported constructors from the three modules:

```go
// 1. Load configuration and build the core module's config managers.
configManagers, err := coresetup.SetupConfigManagers(ctx, applicationYAML, "naryo")

// 2. Set up whichever adapters you want — here, both shipped ones.
persistenceGormModule, err := persistencegormsetup.Setup(ctx, applicationYAML, "naryo")
broadcasterKafkaModule, err := broadcasterkafkasetup.Setup(ctx, applicationYAML, "naryo", configManagers.BroadcasterConfigConfigManager)
broadcasterHttpModule, err := broadcasterhttpsetup.Setup(ctx, configManagers.HttpClientConfigManager, configManagers.BroadcasterConfigConfigManager)

// 3. Wire core against those adapters and start it.
coreModule, err := coresetup.Setup(
    configManagers,
    persistenceGormModule.Stores,
    []broadcast.Producer{broadcasterHttpModule.HttpBroadcasterProducer, broadcasterKafkaModule.KafkaBroadcasterProducer},
)
err = coreModule.Bootstrapper.Start(ctx)
```

(See `server/cmd/server/main.go` for the real, complete version, including migrating the
database and handling each step's error.) You could swap in your own `broadcast.Producer` or
persistence adapter here instead of the shipped ones, or call `coresetup.Setup` with no persistence
at all. It's just Go, wired by hand.

> **This isn't runnable from outside this repo yet.** naryo-go's modules aren't published: their
> module paths (`gitlab.com/iobuilders/projects/eng/naryo-go/...`) point at this private
> repository, so `go get`ting them into an external project doesn't work today. Embedding is
> architecturally supported (this repo's own `server` module is proof, since it's just an importer
> of the other three), but naryo-go needs to be published somewhere resolvable (publicly, or via a
> module proxy your project has access to) before an external Go project can depend on it this way.

## Next Steps

Now that you have naryo-go running, here are some resources to help you continue your journey:

- **[Tutorials](./tutorials/index.md)**: See naryo-go capture and broadcast a real Solana event,
  end-to-end.
- **[Concepts](./concepts.md)**: Learn the vocabulary: Node, Filter, Event, Broadcaster,
  Persistence.
- **[Architecture](./architecture.md)**: Understand naryo-go's layers and functional data flow,
  and how it compares to the Java original.
