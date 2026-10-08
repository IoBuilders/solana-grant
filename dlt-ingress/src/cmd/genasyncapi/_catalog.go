package main

import (
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/event/buildtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/event/custodykey"
	"dlt-ingress/src/main/dltingress/internal/infra/event/signandsend"
)

// The catalog is where every bounded context's event contract is written, and the only place it
// is written. One file, one entry per bounded context, one doc per event: adding an event means
// adding an entry here next to the others, and the gate in resolveMessages makes sure nobody
// forgets. Nothing in src/main knows that AsyncAPI exists.
//
// Only prose and event-sourcing metadata belong here. Channels, addresses, producers, consumers
// and x-orphan are read from the source by discover.go — declaring them again would just be a
// second copy to keep in sync.
//
// A bounded context is registered by appending its spec below.
var catalog = []spec{dltingressSpec}

// ─── dltingress ───────────────────────────────────────────────────────────────────────────────

var dltingressSpec = spec{
	bc:      "dltingress",
	title:   "DLT Ingress Events",
	version: "1.0.0",
	description: `Domain events published by the dltingress bounded context, which owns custody key management,
the transaction dispatch pipeline (build → sign → send → retry) and the faucet wallets that keep
the custody accounts funded.

MODELLING NOTES

The transport is a Postgres outbox (event_stores) drained by an in-process relay, not a message
broker. The servers section is therefore descriptive only, and there are no protocol bindings.
The contract lives in components.messages and components.schemas.

Channels, producers, consumers and x-orphan are read from the code: a channel address is built
from the package that declares the struct, producers from the bus.Publish call sites, consumers
from the listeners registered for the event. An event with no registered consumer anywhere is
marked x-orphan — a finding, not an error: under event sourcing an event with no consumer is
still history.

Structs that embed event.BaseEvent but are never published are excluded as command response DTOs
rather than listed as messages; the generator requires each one to be declared with its reason,
so a genuinely new event cannot pass as one of them.

AsyncAPI has no native concept of an aggregate stream, so event-sourcing metadata is carried in
the custom x-aggregate extension. Whether each aggregate can actually be rebuilt from these
events is NOT expressible here — that analysis lives in docs/events/dltingress-event-sourcing.md.

Payload schemas are reflected from the Go structs, so a field is described here only if its
struct tag describes it. Fields this repository declares carry camelCase json tags; Id and
CreatedAt come from event.BaseEvent in iob-go-core, which has no tags, so they stay PascalCase on
the wire. Amounts are amount.Amount values, serialised as decimal strings that preserve arbitrary
precision and an explicit decimals count; the ones inside the embedded EVM/SVM transaction payload
models reflect as string|null, whether or not the Go field is itself a pointer, while amount fields
declared directly on an event — the faucet wallet ones — reflect as plain strings.`,

	docs: []doc{
		{
			event: custodykey.KeyCreatedEvent{},
			title: "Custody key created",
			summary: "A signing key was created at the custody provider and registered locally, binding a " +
				"dltAccountId to the provider's own externalId.",
			aggregate: aggregate{
				kind:    "CustodyKey",
				idField: "dltAccountId",
				note:    "Natural key. The aggregate Id uuid is not carried — see gap B1.",
			},
			sendNote:    "The custody provider has already returned the key at that point.",
			receiveNote: "That listener republishes it on the cross bus.",
		},
		{
			event:     custodykeyevents.KeyCreatedCrossEvent{},
			title:     "Custody key created (cross-context)",
			summary:   "Field-for-field copy of KeyCreated republished on the cross bus.",
			aggregate: aggregate{kind: "CustodyKey", idField: "dltAccountId"},
		},
		{
			event: transaction.TransactionSentEvent{},
			title: "Transaction signed and submitted",
			summary: "A transaction was signed by the custody provider and submitted to the node. The " +
				"transaction id is deterministic and known before sending, so the row is persisted beforehand.",
			aggregate: aggregate{
				kind:    "EvmTransaction | SvmTransaction",
				idField: "txId",
				note: "Variant selected by the dlt field. The two payload models are embedded pointers in Go " +
					"and flatten into this object on marshal — see gap B6.",
			},
			sendNote: "The transaction has already been submitted to the node at that point.",
		},
		{
			event: transaction.RetriedEvent{},
			title: "Transaction retried",
			summary: "A stuck transaction was re-signed and resubmitted with higher gas, producing a " +
				"replacement transaction. Carries only the two identifiers — see gap B4.",
			aggregate: aggregate{kind: "EvmTransaction", idField: "newTxId"},
		},
		{
			event: failedtransaction.SavedEvent{},
			title: "Failed transaction recorded",
			summary: "A transaction was reported as failed by the blockchain context and recorded for later " +
				"retry. Implies status NOT_RETRIED.",
			aggregate: aggregate{kind: "FailedTransaction", idField: "txId"},
		},
		{
			event: failedtransaction.TransitFailedTransactionToRetriedEvent{},
			title: "Failed transaction moved to retried",
			summary: "The failed transaction was superseded by a retry. Implies status RETRIED. Name is " +
				"imperative where the convention requires past tense — see gap H3.",
			aggregate: aggregate{kind: "FailedTransaction", idField: "txId"},
		},
		{
			event: faucetwallet.CreatedEvent{},
			title: "Faucet wallet created",
			summary: "A faucet wallet was configured for a network: a custody key was provisioned for it and " +
				"bound to the funding amount and the balance threshold that trigger a top-up. Created enabled.",
			aggregate: aggregate{
				kind:    "FaucetWallet",
				idField: "faucetWalletId",
				note: "Both the custody key and the wallet row are already persisted when this is published. " +
					"networkId is a second unique key — there is at most one wallet per network, and that is " +
					"how the handlers look it up.",
			},
		},
		{
			event: faucetwallet.UpdatedEvent{},
			title: "Faucet wallet updated",
			summary: "The funding amount, the balance threshold or the enabled flag of a network's faucet " +
				"wallet was changed. Carries the resulting configuration in full, not a delta.",
			aggregate: aggregate{
				kind:    "FaucetWallet",
				idField: "faucetWalletId",
				note:    "The custody key binding cannot change, so it is not carried here.",
			},
		},
	},

	// These embed event.BaseEvent without being messages: the handler returns the same shape it
	// would have published, and the name came along with it.
	excluded: []exclusion{
		{
			event:  transaction.TransactionBuiltEvent{},
			reason: "buildtransaction command response, returned to the caller and never published",
		},
		{
			event:  signandsendevents.TransactionSentCrossEvent{},
			reason: "signandsend cross-command response, returned by the adapter and never published",
		},
		{
			event:  buildtransactionevents.TransactionBuiltCrossEvent{},
			reason: "buildtransaction cross-command response, returned by the adapter and never published",
		},
	},
}
