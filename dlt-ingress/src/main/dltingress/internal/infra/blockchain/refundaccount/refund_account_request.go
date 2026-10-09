package refundaccount

type Request struct {
	Signer    string `json:"signer"`
	NetworkId string `json:"networkId"`
}
