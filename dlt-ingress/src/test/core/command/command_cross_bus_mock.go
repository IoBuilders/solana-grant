package commandcrossbusmock

import (
	"context"
	"dlt-ingress/src/main/dltingress/config"
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/port/command/signandsend"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/stretchr/testify/mock"
	"github.com/umbracle/ethgo/abi"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

type CrossBusMock struct {
	mock.Mock
}

func (m *CrossBusMock) RegisterPort(port interface{}) error {
	m.Called(port)
	return nil
}

func (m *CrossBusMock) Dispatch(
	ctx context.Context,
	cmd command.CrossCommand,
) (command.CrossResponse, error) {
	args := m.Called(ctx, cmd)

	var response command.CrossResponse
	if args.Get(0) != nil {
		if signAndSendCrossCmd, isSignAndSend := cmd.(*signandsendcross.CrossCommand); isSignAndSend {
			validateSmartContractCall(signAndSendCrossCmd)
		}
		response = args.Get(0).(command.CrossResponse)
	}

	return response, args.Error(1)
}

// As compilation errors no longer occur because we do not use contract wrappers for smart contract calls, this function is for validating smart contract calls are done correctly.
func validateSmartContractCall(signAndSendCrossCmd *signandsendcross.CrossCommand) {
	var valError error
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("invalid smart contract call: invalid method args for method '%s' of the smart contract '%s': %w",
				signAndSendCrossCmd.MethodName, signAndSendCrossCmd.SmartContractName, r.(error)))
		}
		if valError != nil {
			panic(valError)
		}
	}()
	abiString, ok := dltingressconfig.SmartContractDefinitions[common.EVM][signAndSendCrossCmd.SmartContractName]
	if !ok {
		valError = fmt.Errorf("invalid smart contract call: smart contract '%s' not found in smart_contract_definitions.go",
			signAndSendCrossCmd.SmartContractName)
		return
	}
	contractAbi, err := abi.NewABI(abiString)
	if err != nil {
		valError = fmt.Errorf("invalid smart contract call: error parsing ABI for smart contract '%s'", signAndSendCrossCmd.SmartContractName)
		return
	}
	contractMethod := contractAbi.GetMethod(signAndSendCrossCmd.MethodName)
	if contractMethod == nil {
		valError = fmt.Errorf("invalid smart contract call: method '%s' not found for smart contract '%s'", signAndSendCrossCmd.MethodName,
			signAndSendCrossCmd.SmartContractName)
		return
	}
	// Exclude hex.ErrLength, "0x prefix not found", "is not correct, expected 20" and "encoding/hex: invalid byte:" because tests use invalid hex addresses
	if _, err := contractMethod.Encode(signAndSendCrossCmd.MethodArgs); err != nil && !errors.Is(err, hex.ErrLength) && err.Error() != "0x prefix not found" &&
		!strings.Contains(err.Error(), "is not correct, expected 20") && !strings.Contains(err.Error(), "encoding/hex: invalid byte:") {
		valError = fmt.Errorf("invalid smart contract call: invalid method args for method '%s' of the smart contract '%s': %s",
			signAndSendCrossCmd.MethodName, signAndSendCrossCmd.SmartContractName, err.Error())
	}
}
