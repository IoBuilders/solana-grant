package eventstorepersistence

func Models() []any {
	return []any{
		&SolanaTransactionEvent{},
		&SolanaInstruction{},
		&SolanaInstructionAccount{},
		&SolanaTransactionAccount{},
		&SolanaTransactionLog{},
		&SolanaBlockEvent{},
		&SolanaBlockTransaction{},
		&SolanaBlockInstruction{},
		&SolanaBlockInstructionAccount{},
		&SolanaBlockTransactionAccount{},
		&SolanaBlockTransactionLog{},
		&SolanaContractEvent{},
		&ContractEventParameter{},
		&SolanaLatestBlock{},
	}
}
