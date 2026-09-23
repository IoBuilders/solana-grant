package contractreadcross

import "math/big"

type CrossQuery struct {
	SmartContractId   string
	SmartContractName string
	MethodName        string
	MethodArgs        map[string]any
	NetworkId         string
	BlockNumber       *big.Int
}

type Response struct {
	Result map[string]any
}
