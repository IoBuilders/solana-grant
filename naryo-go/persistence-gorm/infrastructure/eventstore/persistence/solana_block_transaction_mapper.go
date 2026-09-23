package eventstorepersistence

import (
	"slices"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func FromSolanaTransactionDomain(t event.SolanaTransaction) (*SolanaBlockTransaction, error) {
	instructions := make([]SolanaBlockInstruction, 0, len(t.Instructions))
	for index, instruction := range t.Instructions {
		accounts := make([]SolanaBlockInstructionAccount, 0, len(instruction.Accounts))
		for accountIndex, address := range instruction.Accounts {
			accounts = append(accounts, SolanaBlockInstructionAccount{
				Index:   accountIndex,
				Address: address,
			})
		}

		instructions = append(instructions, SolanaBlockInstruction{
			Index:     index,
			ProgramID: instruction.ProgramID,
			Data:      instruction.Data,
			Accounts:  accounts,
			Inner:     instruction.Inner,
		})
	}

	accounts := make([]SolanaBlockTransactionAccount, 0, len(t.Accounts))
	for index, address := range t.Accounts {
		accounts = append(accounts, SolanaBlockTransactionAccount{Index: index, Address: address})
	}

	logs := make([]SolanaBlockTransactionLog, 0, len(t.Logs))
	for index, message := range t.Logs {
		logs = append(logs, SolanaBlockTransactionLog{Index: index, Message: message})
	}

	return &SolanaBlockTransaction{
		TransactionID: t.Id(),
		NodeID:        t.NodeID(),
		Err:           t.Err,
		Slot:          t.Slot,
		Instructions:  instructions,
		Accounts:      accounts,
		Logs:          logs,
	}, nil
}

// ToSolanaTransactionDomain orders the owned rows by their stored index
// rather than trusting the order they were loaded in: the index is what
// carries the meaning, and a query without an ORDER BY makes no promise.
func ToSolanaTransactionDomain(m SolanaBlockTransaction) (event.SolanaTransaction, error) {
	storedInstructions := slices.Clone(m.Instructions)
	slices.SortFunc(storedInstructions, func(a, b SolanaBlockInstruction) int {
		return a.Index - b.Index
	})

	instructions := make([]event.InstructionData, 0, len(storedInstructions))
	for _, instruction := range storedInstructions {
		storedInstructionAccounts := slices.Clone(instruction.Accounts)
		slices.SortFunc(storedInstructionAccounts, func(a, b SolanaBlockInstructionAccount) int {
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
	slices.SortFunc(storedAccounts, func(a, b SolanaBlockTransactionAccount) int {
		return a.Index - b.Index
	})

	accounts := make([]string, 0, len(storedAccounts))
	for _, account := range storedAccounts {
		accounts = append(accounts, account.Address)
	}

	storedLogs := slices.Clone(m.Logs)
	slices.SortFunc(storedLogs, func(a, b SolanaBlockTransactionLog) int {
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

	return event.NewSolanaTransaction(
		m.NodeID,
		m.TransactionID,
		m.Slot,
		instructions,
		accounts,
		logs,
		errPayload,
	)
}
