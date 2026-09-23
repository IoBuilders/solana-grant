# Concepts

This page explains the core concepts you will encounter when using naryo-go.

## Node

Represents a connection to a blockchain endpoint: a Solana RPC node. A deployment may define
multiple nodes.

## Filter

A filter describes which on-chain data should be captured. Two kinds exist:

- **Event filter**: matches contract events, by scope (e.g. an Anchor program's `EMIT_CPI`
  events) and a chain-specific specification (program address + event signature).
- **Transaction filter**: matches transactions by identifier and status.

Filters are added at startup via configuration.

## Event

The normalized representation of a captured occurrence that matched a filter. It may include:

- Event type (`CONTRACT`, `TRANSACTION`, `BLOCK`)
- Node ID
- Parameters (decoded arguments for contract events)
- Transaction and block context (signatures, slot numbers)
- Status: for contract events, the Solana commitment level it was observed at
  (`PROCESSED`/`CONFIRMED`/`FINALIZED`); a filter can constrain which levels it matches

## Broadcaster

Responsible for delivering events to a destination. naryo-go currently ships an HTTP broadcaster
(`broadcaster-http`). A broadcaster references a destination configuration and a target: the
kind of event (contract event or transaction) routed to a given path.

## Persistence

Optional storage for durability and replay. naryo-go currently ships a GORM/Postgres integration
(`persistence-gorm`). When configured, ingested block/transaction/contract-event data and filter
sync state are stored in the database.

## Runtime model

At runtime, naryo-go:

1. Connects to configured nodes
2. Polls for new slots
3. Applies filters and decodes data into events
4. Optionally persists matched events
5. Broadcasts matched events to one or more destinations

For configuration details see
[`examples/quickstart/application.yaml`](../examples/quickstart/application.yaml).

## Compared to the Java original

naryo-go's contract event `Status` reflects Solana's own commitment levels
(`PROCESSED`/`CONFIRMED`/`FINALIZED`), set once when the event is captured. It's not the same as
the Java original's **Confirmation and invalidation** concept, which counts block confirmations
after the fact and can invalidate an already-broadcast event if a reorg drops it. naryo-go
doesn't retroactively revisit or invalidate an event once emitted. See
[Architecture](architecture.md#compared-to-the-java-original) for the full list of what's ported
so far.
