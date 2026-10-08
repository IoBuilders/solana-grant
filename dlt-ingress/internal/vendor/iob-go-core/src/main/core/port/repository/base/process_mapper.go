package baserepo

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

func ToDomain(model Process) base.Process {
	var unsignedTx *base.UnsignedTransactionProcess
	if model.UnsignedTransactionProcess != nil {
		u := base.UnsignedTransactionProcess{
			Id:           model.UnsignedTransactionProcess.Id,
			ProcessId:    model.UnsignedTransactionProcess.ProcessId,
			DltAccountId: model.UnsignedTransactionProcess.DltAccountId,
			Payload:      model.UnsignedTransactionProcess.Payload,
		}
		unsignedTx = &u
	}
	var subprocesses []base.Process
	for _, subprocess := range model.Subprocesses {
		subprocesses = append(subprocesses, ToDomain(subprocess))
	}
	process := base.NewProcess()
	process.Entity = base.Entity{
		Id:        model.Model.Id,
		CreatedAt: model.Model.CreatedAt,
		UpdatedAt: model.Model.UpdatedAt,
	}
	process.ParentId = model.ParentId
	process.Status = model.Status
	process.Type = model.Type
	process.SigningType = model.SigningType
	process.EntityId = model.EntityId
	process.TransactionHash = model.TransactionHash
	process.SignerDltAccountId = model.SignerDltAccountId
	process.RevertReason = model.RevertReason
	process.Callback = model.Callback
	process.UnsignedTransactionProcess = unsignedTx
	process.Subprocesses = subprocesses
	return process
}

func FromDomain(domain base.Process) Process {
	var unsignedTx *UnsignedTransactionProcess
	if domain.UnsignedTransactionProcess != nil {
		u := UnsignedTransactionProcess{
			Id:           domain.UnsignedTransactionProcess.Id,
			ProcessId:    domain.UnsignedTransactionProcess.ProcessId,
			DltAccountId: domain.UnsignedTransactionProcess.DltAccountId,
			Payload:      domain.UnsignedTransactionProcess.Payload,
		}
		unsignedTx = &u
	}
	var subprocesses []Process
	for _, subprocess := range domain.Subprocesses {
		subprocesses = append(subprocesses, FromDomain(subprocess))
	}
	return Process{
		Model: Model{
			Id:        domain.Entity.Id,
			CreatedAt: domain.Entity.CreatedAt,
			UpdatedAt: domain.Entity.UpdatedAt,
		},
		ParentId:                   domain.ParentId,
		Status:                     domain.Status,
		Type:                       domain.Type,
		SigningType:                domain.SigningType,
		EntityId:                   domain.EntityId,
		TransactionHash:            domain.TransactionHash,
		SignerDltAccountId:         domain.SignerDltAccountId,
		RevertReason:               domain.RevertReason,
		Callback:                   domain.Callback,
		UnsignedTransactionProcess: unsignedTx,
		Subprocesses:               subprocesses,
	}
}
