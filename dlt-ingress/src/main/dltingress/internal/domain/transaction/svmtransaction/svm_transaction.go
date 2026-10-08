package svmtransaction

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/validate"
)

// MaxSvmTransactionSize is Solana's wire transaction size limit (bytes).
const MaxSvmTransactionSize = 1232

type SvmTransaction struct {
	transaction.Transaction
	Dlt                   common.Dlt
	FeePayer              string
	RecentBlockhash       string
	CuLimit               *amount.Amount
	CuPrice               *amount.Amount
	SerializedTransaction string
}

func NewSvmTransaction(
	txId string,
	networkId string,
	networkUrl string,
	feePayer string,
	recentBlockhash string,
	serializedTransaction string,
	cuLimit *amount.Amount,
	cuPrice *amount.Amount,
) (*SvmTransaction, error) {
	if err := validate.StringMaxLength(txId, "TxId", "SvmTransaction", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(networkId, "NetworkId", "SvmTransaction", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(networkUrl, "NetworkUrl", "SvmTransaction", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(feePayer, "FeePayer", "SvmTransaction", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(recentBlockhash, "RecentBlockhash", "SvmTransaction", 100); err != nil {
		return nil, err
	}

	return &SvmTransaction{
		Transaction: transaction.Transaction{
			TxId:       txId,
			NetworkId:  networkId,
			NetworkUrl: networkUrl,
		},
		Dlt:                   common.SVM,
		FeePayer:              feePayer,
		RecentBlockhash:       recentBlockhash,
		SerializedTransaction: serializedTransaction,
		CuLimit:               cuLimit,
		CuPrice:               cuPrice,
	}, nil
}

func (tx *SvmTransaction) Clone() *SvmTransaction {
	if tx == nil {
		return nil
	}

	return &SvmTransaction{
		Transaction:           tx.Transaction,
		Dlt:                   tx.Dlt,
		FeePayer:              tx.FeePayer,
		RecentBlockhash:       tx.RecentBlockhash,
		SerializedTransaction: tx.SerializedTransaction,
		CuLimit:               tx.CuLimit,
		CuPrice:               tx.CuPrice,
	}
}
