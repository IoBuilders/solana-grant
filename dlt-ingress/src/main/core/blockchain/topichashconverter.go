package blockchain

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

func ConvertSignatureToTopicHash(signature string) common.Hash {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(signature))
	return common.BytesToHash(hash.Sum(nil))
}

func GetEventSignatureFromABI(abiString string, eventName string) string {
	parsedABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		panic(err)
	}
	event, exists := parsedABI.Events[eventName]
	if !exists {
		panic(fmt.Sprintf("event %s doesn not exist", eventName))
	}
	return event.Sig
}

func GetEventTopicHashFromABI(abiString string, eventName string) string {
	return ConvertSignatureToTopicHash(GetEventSignatureFromABI(abiString, eventName)).String()
}

func IsKeccak256Hash(s string) bool {
	if !strings.HasPrefix(s, "0x") || len(s) != 66 {
		return false
	}
	_, err := hex.DecodeString(s[2:])
	return err == nil
}
