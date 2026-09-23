package eventstorepersistence

import (
	"slices"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func FromSolanaTransactionEventDomain(e event.SolanaTransactionEvent) (*SolanaTransactionEvent, error) {
	instructions := make([]SolanaInstruction, 0, len(e.Instructions))
	for index, instruction := range e.Instructions {
		accounts := make([]SolanaInstructionAccount, 0, len(instruction.Accounts))
		for accountIndex, address := range instruction.Accounts {
			accounts = append(accounts, SolanaInstructionAccount{
				Index:   accountIndex,
				Address: address,
			})
		}

		instructions = append(instructions, SolanaInstruction{
			Index:     index,
			ProgramID: instruction.ProgramID,
			Data:      instruction.Data,
			Accounts:  accounts,
			Inner:     instruction.Inner,
		})
	}

	accounts := make([]SolanaTransactionAccount, 0, len(e.Accounts))
	for index, address := range e.Accounts {
		accounts = append(accounts, SolanaTransactionAccount{Index: index, Address: address})
	}

	logs := make([]SolanaTransactionLog, 0, len(e.Logs))
	for index, message := range e.Logs {
		logs = append(logs, SolanaTransactionLog{Index: index, Message: message})
	}

	return &SolanaTransactionEvent{
		TransactionID: e.Id(),
		NodeID:        e.NodeID(),
		Err:           e.Err,
		Slot:          e.Slot,
		Instructions:  instructions,
		Accounts:      accounts,
		Logs:          logs,
	}, nil
}

// ToTransactionEvent orders the owned rows by their stored index rather than
// trusting the order they were loaded in: the index is what carries the meaning,
// and a query without an ORDER BY makes no promise.
func ToTransactionEventDomain(m SolanaTransactionEvent) (event.TransactionEvent, error) {
	storedInstructions := slices.Clone(m.Instructions)
	slices.SortFunc(storedInstructions, func(a, b SolanaInstruction) int {
		return a.Index - b.Index
	})

	instructions := make([]event.InstructionData, 0, len(storedInstructions))
	for _, instruction := range storedInstructions {
		storedInstructionAccounts := slices.Clone(instruction.Accounts)
		slices.SortFunc(storedInstructionAccounts, func(a, b SolanaInstructionAccount) int {
			return a.Index - b.Index
		})

		instructionAccounts := make([]string, 0, len(storedInstructionAccounts))
		for _, account := range storedInstructionAccounts {
			instructionAccounts = append(instructionAccounts, account.Address)
		}

		instructions = append(instructions, event.InstructionData{
			ProgramID: instruction.ProgramID,
			Data:      instruction.Data,
			Accounts:  instructionAccounts,
			Inner:     instruction.Inner,
		})
	}

	storedAccounts := slices.Clone(m.Accounts)
	slices.SortFunc(storedAccounts, func(a, b SolanaTransactionAccount) int {
		return a.Index - b.Index
	})

	accounts := make([]string, 0, len(storedAccounts))
	for _, account := range storedAccounts {
		accounts = append(accounts, account.Address)
	}

	storedLogs := slices.Clone(m.Logs)
	slices.SortFunc(storedLogs, func(a, b SolanaTransactionLog) int {
		return a.Index - b.Index
	})

	logs := make([]string, 0, len(storedLogs))
	for _, log := range storedLogs {
		logs = append(logs, log.Message)
	}

	var errPayload *string
	if m.Err != nil {
		errPayload = m.Err
	}

	return event.NewSolanaTransactionEvent(
		m.NodeID,
		m.TransactionID,
		m.Slot,
		instructions,
		accounts,
		logs,
		errPayload,
	)
}
