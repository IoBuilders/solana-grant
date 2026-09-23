# naryo-go

naryo-go is a Go port of [Naryo](https://github.com/LF-Decentralized-Trust-labs/Naryo), a
lightweight, modular framework for capturing, processing, and broadcasting events from
Distributed Ledger Technology (DLT) networks. Whether you're building a dApp, a data analytics
pipeline, or a monitoring tool, naryo-go simplifies blockchain integration so you can focus on
your application's core logic.

**naryo-go is Solana-only for now.** The Java original also supports Ethereum-compatible (EVM)
chains and Hedera, plus a few features naryo-go hasn't ported yet (Kafka/RabbitMQ broadcasting,
MongoDB persistence, confirmation/invalidation on reorgs, a dynamic configuration API). See
[Architecture § Compared to the Java original](docs/architecture.md#compared-to-the-java-original)
for the full list.

## Table of Contents

- [Why naryo-go?](#why-naryo-go)
- [Architecture](#architecture)
- [Quickstart](#quickstart)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Modules](#modules)
- [Contributing](#contributing)
- [License](#license)

## Why naryo-go?

- **Solana Support**: Connect to a Solana RPC node and capture contract events and transactions.
- **HTTP Broadcasting**: Broadcast events to your systems over HTTP.
- **Framework-Agnostic Core**: The heart of naryo-go is a framework-agnostic core, kept free of
  any specific transport or database.
- **Powerful Filtering**: Define granular filters for contract events (by program + Anchor event
  signature) and transactions (by status).
- **Optional Persistence**: Store ingested chain data and filter sync state in Postgres for
  durability and replay.

## Architecture

For a deeper dive into naryo-go's architecture and core concepts, check out our
[architecture](docs/architecture.md) and [concepts](docs/concepts.md) documentation.

## Quickstart

Want to see naryo-go in action? Our Docker-based quickstart will get you up and running in
minutes.

- [Naryo-go Quickstart: Solana](docs/tutorials/naryo_quickstart.md)

## Getting Started

Run naryo-go as-is via the reference `server` binary, or embed `core`, `broadcaster-http`,
`broadcaster-kafka`, and `persistence-gorm` directly into your own Go program the way `server`
itself does. See
[Getting Started](docs/getting_started.md) for both, including the one thing that currently blocks
embedding from outside this repo (naryo-go's modules aren't published anywhere yet).

## Configuration

naryo-go is configured via a single YAML file merged over `server`'s embedded default. See
[`examples/quickstart/application.yaml`](examples/quickstart/application.yaml) for a complete,
annotated example covering nodes, filters, broadcasting, and store configuration. naryo-go doesn't
have a dedicated configuration reference doc yet the way the Java original does. See
[Getting Started § Running naryo-go with Docker](docs/getting_started.md#1-running-naryo-go-with-docker)
for pointers to the source of truth in the meantime.

## Modules

naryo-go is split into libraries `server` composes, mirroring the Java original's modular design,
though unlike Java, there's currently only one broadcaster and one persistence adapter to choose
from.

- **core**: The framework-agnostic runtime, domain model, and application ports.
- **broadcaster-http**: HTTP broadcaster adapter.
- **broadcaster-kafka**: Kafka broadcaster adapter.
- **persistence-gorm**: GORM/Postgres persistence adapter, with Atlas-managed migrations.
- **server**: The reference microservice binary that wires `core`, `broadcaster-http`, and
  `persistence-gorm` together. The only runnable artifact in this repo.


## License

This project is licensed under the [Apache License, Version 2.0](LICENSE).
