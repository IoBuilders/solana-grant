package savefailedtransaction

type Request struct {
	TransactionHash string `json:"transactionHash"`
	RevertReason    string `json:"revertReason,omitempty"`
	NetworkId       string `json:"networkId"`
}
