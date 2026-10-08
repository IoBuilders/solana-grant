package custodykeyevents

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"

type KeyCreatedCrossEvent struct {
	event.BaseEvent
	DltAccountId    string `json:"dltAccountId" required:"true" description:"Blockchain address the key signs for, and the natural key of the custody key."`
	ExternalId      string `json:"externalId" required:"true" description:"Identifier of the key at the custody provider."`
	Dlt             string `json:"dlt" required:"true" enum:"EVM,SVM,Hashgraph" description:"Widened from common.Dlt - replay must re-parse via ParseDlt."`
	CustodyProvider string `json:"custodyProvider" required:"true" enum:"DFNS,KMS" description:"Service holding the private key."`
}
