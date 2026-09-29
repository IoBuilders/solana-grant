package eventstore

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
)

func TestCoreEventStore_NewEventStore_Success(t *testing.T) {
	eventType := "type"
	payload, _ := json.Marshal(map[string]interface{}{
		"key": "value",
	})
	consumer, _ := NewEventConsumer("type", string(Pending), false)
	consumers := []EventConsumer{*consumer}
	traceParent := "traceparent"
	publicationType := Memory
	eventStore, err := NewEventStore(eventType, payload, consumers, traceParent, string(publicationType), false)

	assert.Nil(t, err)
	assert.NotNil(t, eventStore)
	assert.Equal(t, eventStore.Type, eventType)
	var expectedPayload, realPayload map[string]interface{}
	_ = json.Unmarshal(payload, &expectedPayload)
	_ = json.Unmarshal(eventStore.Payload, &realPayload)
	assert.Equal(t, expectedPayload, realPayload)
	assert.Equal(t, eventStore.EventConsumers, consumers)
	assert.Equal(t, eventStore.TraceParent, traceParent)
	assert.Equal(t, eventStore.PublicationType, publicationType)
}

func TestCoreEventStore_NewEventStore_Type_Empty_Error(t *testing.T) {
	payload, _ := json.Marshal(make(map[string]interface{}))
	eventStore, err := NewEventStore("", payload, []EventConsumer{}, "", string(Memory), false)

	assert.Nil(t, eventStore)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeValidation)
}

func TestCoreEventStore_NewEventStore_Type_TooLong_Error(t *testing.T) {
	payload, _ := json.Marshal(make(map[string]interface{}))
	eventStore, err := NewEventStore(strings.Repeat("1", 256), payload, []EventConsumer{}, "", string(Memory), false)

	assert.Nil(t, eventStore)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeValidation)
}

func TestCoreEventStore_NewEventStore_Payload_Nil_Error(t *testing.T) {
	eventStore, err := NewEventStore("type", nil, []EventConsumer{}, "", string(Memory), false)

	assert.Nil(t, eventStore)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeEventStoreEmptyPayload)
}

func TestCoreEventStore_NewEventStore_Payload_Empty_Error(t *testing.T) {
	payload, _ := json.Marshal(make(map[string]interface{}))
	eventStore, err := NewEventStore("type", payload, []EventConsumer{}, string(Memory), string(Memory), false)

	assert.Nil(t, eventStore)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeEventStoreEmptyPayload)
}

func TestCoreEventStore_NewEventStore_Invalid_PublicationType_Error(t *testing.T) {
	publicationType := "invalid"
	eventType := "type"
	payload, _ := json.Marshal(map[string]interface{}{
		"key": "value",
	})
	consumer, _ := NewEventConsumer("type", string(Pending), false)
	consumers := []EventConsumer{*consumer}
	eventStore, err := NewEventStore(eventType, payload, consumers, "", publicationType, false)

	assert.Nil(t, eventStore)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeEventStoreInvalidPublicationType)
}
