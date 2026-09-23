package clientblockchain

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Service interface {
	GetClient() *ethclient.Client
}

type ServiceImpl struct {
	Client *ethclient.Client
}

func (b *ServiceImpl) GetClient() *ethclient.Client {
	return b.Client
}

func NewBlockchainService(blockchainClientUrl string) *ServiceImpl {
	client, err := ethclient.Dial(blockchainClientUrl)
	if err != nil {
		panic(fmt.Errorf("blockchain client connection failed: %fv", err))
	}
	return &ServiceImpl{Client: client}
}
