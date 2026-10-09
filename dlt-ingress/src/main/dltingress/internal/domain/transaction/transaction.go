package transaction

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

type Transaction struct {
	base.Entity
	TxId         string
	OriginalTxId *string
	NetworkId    string
	NetworkUrl   string
}
