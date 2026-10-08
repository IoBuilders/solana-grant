package contractread

import "math/big"

type Query struct {
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
