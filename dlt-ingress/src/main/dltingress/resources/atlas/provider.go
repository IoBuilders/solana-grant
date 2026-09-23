//go:build gorm

package main

import (
	"dlt-ingress/src/main/dltingress/domain/custodykey"
	"dlt-ingress/src/main/dltingress/domain/nonce"
	"dlt-ingress/src/main/dltingress/domain/queuelock"
	"dlt-ingress/src/main/dltingress/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/domain/txqueueslot"
	"fmt"

	"ariga.io/atlas-provider-gorm/gormschema"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

func main() {
	schema, err := GetSchema()
	if err != nil {
		logger.Error("Schema generation failed", "error", err)
		return
	}
	fmt.Println(schema)
}

func Models() []any {
	models := []any{
		&custodykey.CustodyKey{},
		&eventstore.EventConsumer{},
		&eventstore.EventStore{},
		&nonce.Nonce{},
		&evmtransaction.EvmTransaction{},
		&svmtransaction.SvmTransaction{},
		&queuelock.QueueLock{},
		&txqueueslot.TxQueueSlot{},
		&failedtransaction.FailedTransaction{},
	}
	return models
}

func GetSchema() (string, error) {
	models := Models()
	if len(models) == 0 {
		return "", fmt.Errorf("no models found")
	}

	schema, err := gormschema.New("postgres").Load(models...)
	if err != nil {
		return "", fmt.Errorf("failed to load schema: %v", err)
	}

	return schema, nil
}
