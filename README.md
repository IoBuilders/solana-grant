# Grant

## Table of Contents

1. [Context](#1-context)
   - [ML1: Programs](#ml1-programs)
   - [ML2: Blockchain Gateway](#ml2-blockchain-gateway)
   - [ML3: Blockchain Listener](#ml3-blockchain-listener)
2. [Delivery](#2-delivery)
   - [ML1](#ml1)
   - [ML2](#ml2)
   - [ML3](#ml3)

## 1. Context

This repository documents the deliverables produced by IOBuilders under the Grant Agreement with the Solana Foundation, awarded to support the development of Solana-based components for **Asseto**, ioBuilders' platform for tokenized financial instruments.

The agreement defines three deliverables, each corresponding to a milestone tracked in this repository:

| Milestone | Deliverable | Folder | Status |
|---|---|---|---|
| ML1 | Solana programs for a financial instrument | [`asseto-solana-programs/`](asseto-solana-programs/) | Approved |
| ML2 | Blockchain Gateway extension for Solana | [`dlt-ingress/`](dlt-ingress/) | Ready for review |
| ML3 | Naryo (DLT listener) extension for Solana | [`naryo-go/`](naryo-go/) | Ready for review |

Each milestone folder contains its own documentation describing the deliverable itself (architecture, functionality, how to verify it). This document explains what was agreed for each milestone and why, and, in [Section 2](#2-delivery), how each one is justified as complete.

This repository is scoped solely to evidencing the agreed milestones for grant acceptance purposes. It is not intended for long-term maintenance — ongoing iteration on each component continues in ioBuilders' private repositories.

### ML1: Programs

Re-implementation of the core Asseto smart contracts as Solana programs (Rust), covering the full lifecycle of a financial instrument. The instrument delivered under this milestone is a **Bond**.

Capabilities include:

- **Core token operations** — metadata management, transfers, minting, and burning (issuer-initiated burn works even on frozen balances).
- **Global control and state management** — pause/unpause, and a permanent, irreversible deactivation that preserves historical balances on-chain.
- **Holder-specific controls** — full or partial address freezing, limiting transfer capability while still allowing incoming transfers.
- **Transfer control mechanism** — selectable restriction modes (No Restriction, Clearing, Whitelist), with issuer-controlled whitelist management, mode switching, and restriction removal.
- **Coupon and snapshot capability** — issuer-triggered, verifiable on-chain snapshots of holder balances, used for dividend calculation, voting eligibility, or regulatory reporting.

Development of the equivalent Fund smart contracts is progressing in parallel as part of ioBuilders' internal roadmap, but is outside the scope of this grant-funded deliverable.

### ML2: Blockchain Gateway

Extension of the Asseto Blockchain Gateway to support seamless interaction with Solana programs, enabling ioBuilders' off-chain services to operate on Solana in the same way they currently do on EVM and Hedera.

### ML3: Blockchain Listener

Naryo, ioBuilders' open-source DLT event listener, already exists as a Java implementation. Extending that existing Java version to support Solana is out of scope for this grant; instead, this deliverable is a new, separate implementation of Naryo in Go, adding Solana support to enable real-time monitoring of Asseto-related on-chain activity.

This deliverable is scoped to a functional integration for Asseto. Porting this Solana support back into the Java version, to open-source it for the broader Naryo community, is part of ioBuilders' plans but falls outside the scope of this grant-funded deliverable.

## 2. Delivery

This section states, for each milestone, why the delivered work is considered complete against what was agreed in [Section 1](#1-context).

Beyond each milestone individually, [TESTING.md](TESTING.md) walks through running all three together: dlt-ingress invoking the Solana programs, and naryo-go detecting and broadcasting the resulting on-chain event.

### ML1

All capabilities described in [ML1: Programs](#ml1-programs) are implemented and covered by tests. The design and architecture behind the implementation are documented as well, and follow the good practices and standards of the Solana/Anchor ecosystem. The corresponding code is included in the [`asseto-solana-programs/`](asseto-solana-programs/) folder.

### ML2

The requirement is satisfied: the Asseto Blockchain Gateway can build, sign, and send SVM transactions using the same component ioBuilders already uses for EVM, giving off-chain services the same Solana interaction it provides for EVM and Hedera. Further details are included in the [`dlt-ingress/`](dlt-ingress/) folder.

### ML3

Naryo Go has been implemented and is functional, covering an initial feature set:

- Event detection through block retrieval and processing via new-slots polling.
- Usage of Event Filters (with signatures) to detect the program events the app is interested in.
- Usage of Transaction Filters to detect interesting transactions (most often configured for failed transactions).
- Integration with a minimal broadcasting mechanism to let other components (Asseto) know about the events.
- The app resumes from the latest processed block when restarted or when problems occur.
- The solution follows the existing Naryo Java version's architecture and principles, developed in Go.

Further iteration is planned to introduce Geyser streams and other features that improve scalability and resilience. Further details are included in the [`naryo-go/`](naryo-go/) folder.
