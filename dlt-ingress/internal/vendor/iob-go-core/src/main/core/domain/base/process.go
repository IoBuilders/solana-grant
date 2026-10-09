package base

import (
	"sync"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/callback"
)

type ProcessStatus string

const (
	Started                   ProcessStatus = "STARTED"
	Ordered                   ProcessStatus = "ORDERED"
	Created                   ProcessStatus = "CREATED"
	Finished                  ProcessStatus = "FINISHED"
	Failed                    ProcessStatus = "FAILED"
	TransactionBuilt          ProcessStatus = "TRANSACTION_BUILT"
	AwaitingSignedTransaction ProcessStatus = "AWAITING_SIGNED_TRANSACTION"
)

var (
	validStatus = map[ProcessStatus]struct{}{
		Started:                   {},
		Ordered:                   {},
		Created:                   {},
		Finished:                  {},
		Failed:                    {},
		TransactionBuilt:          {},
		AwaitingSignedTransaction: {},
	}
)

func RegisterProcessStatus(statuses ...ProcessStatus) {
	for _, status := range statuses {
		validStatus[status] = struct{}{}
	}
}

func (t ProcessStatus) IsValid() bool {
	_, ok := validStatus[t]
	return ok
}

type Type string

type SigningType string

const (
	Custodial    SigningType = "CUSTODIAL"
	NonCustodial SigningType = "NON_CUSTODIAL"
)

type Process struct {
	Entity
	ParentId                   *uuid.UUID
	Status                     ProcessStatus
	Type                       Type
	SigningType                SigningType
	EntityId                   uuid.UUID
	TransactionHash            string
	SignerDltAccountId         string
	RevertReason               string
	Callback                   callback.Callback
	UnsignedTransactionProcess *UnsignedTransactionProcess
	Subprocesses               []Process

	mu *sync.RWMutex
}

func NewProcess() Process {
	return Process{
		Entity: NewEntity(),
		mu:     &sync.RWMutex{},
	}
}

func NewSubprocess(parentId uuid.UUID, processType Type) Process {
	process := NewProcess()
	process.ParentId = &parentId
	process.Type = processType
	process.Status = Started
	return process
}

func (p *Process) IsSubprocess() bool {
	return p.ParentId != nil
}

func (p *Process) ParentProcessId() uuid.UUID {
	if p.ParentId == nil {
		return uuid.Nil
	}
	return *p.ParentId
}

func (p *Process) MarkStatus(status ProcessStatus) bool {
	return p.markStatus(status, "")
}

func (p *Process) MarkFinished() bool {
	return p.markStatus(Finished, "")
}

func (p *Process) MarkFailed(revertReason string) bool {
	return p.markStatus(Failed, revertReason)
}

func (p *Process) RegisterSubprocess(processId uuid.UUID, subprocessType Type) bool {
	p.mutex().Lock()
	defer p.mutex().Unlock()
	return p.registerSubprocess(processId, subprocessType, Started)
}

func (p *Process) MarkSubprocessStatus(processId uuid.UUID, status ProcessStatus) bool {
	return p.markSubprocess(processId, status, "")
}

func (p *Process) MarkSubprocessFailed(processId uuid.UUID, revertReason string) bool {
	return p.markSubprocess(processId, Failed, revertReason)
}

func (p *Process) UpsertSubprocessStatus(processId uuid.UUID, subprocessType Type, status ProcessStatus) bool {
	if !status.IsValid() {
		return false
	}
	p.mutex().Lock()
	defer p.mutex().Unlock()
	if p.setSubprocessStatus(processId, status, "") {
		return true
	}
	return p.registerSubprocess(processId, subprocessType, status)
}

func (p *Process) Subprocess(processId uuid.UUID) (Process, bool) {
	p.mutex().RLock()
	defer p.mutex().RUnlock()
	index := p.indexOfSubprocess(processId)
	if index < 0 {
		return Process{}, false
	}
	return p.Subprocesses[index], true
}

func (p *Process) SubprocessByType(subprocessType Type) (Process, bool) {
	p.mutex().RLock()
	defer p.mutex().RUnlock()
	for i := range p.Subprocesses {
		if p.Subprocesses[i].Type == subprocessType {
			return p.Subprocesses[i], true
		}
	}
	return Process{}, false
}

func (p *Process) SubprocessesByType(subprocessType Type) []Process {
	p.mutex().RLock()
	defer p.mutex().RUnlock()
	var result []Process
	for i := range p.Subprocesses {
		if p.Subprocesses[i].Type == subprocessType {
			result = append(result, p.Subprocesses[i])
		}
	}
	return result
}

func (p *Process) AllSubprocessesFinished() bool {
	p.mutex().RLock()
	defer p.mutex().RUnlock()
	if len(p.Subprocesses) == 0 {
		return false
	}
	for i := range p.Subprocesses {
		if p.Subprocesses[i].Status != Finished {
			return false
		}
	}
	return true
}

func (p *Process) AnySubprocessFailed() bool {
	p.mutex().RLock()
	defer p.mutex().RUnlock()
	for i := range p.Subprocesses {
		if p.Subprocesses[i].Status == Failed {
			return true
		}
	}
	return false
}

func (p *Process) markStatus(status ProcessStatus, revertReason string) bool {
	if !status.IsValid() {
		return false
	}
	p.mutex().Lock()
	defer p.mutex().Unlock()
	p.Status = status
	if revertReason != "" {
		p.RevertReason = revertReason
	}
	return true
}

func (p *Process) markSubprocess(processId uuid.UUID, status ProcessStatus, revertReason string) bool {
	if !status.IsValid() {
		return false
	}
	p.mutex().Lock()
	defer p.mutex().Unlock()
	return p.setSubprocessStatus(processId, status, revertReason)
}

func (p *Process) mutex() *sync.RWMutex {
	if p.mu == nil {
		p.mu = &sync.RWMutex{}
	}
	return p.mu
}

func (p *Process) registerSubprocess(processId uuid.UUID, subprocessType Type, status ProcessStatus) bool {
	if processId == uuid.Nil || processId == p.Id || p.indexOfSubprocess(processId) >= 0 {
		return false
	}
	subprocess := NewSubprocess(p.Id, subprocessType)
	subprocess.Id = processId
	subprocess.Status = status
	subprocess.EntityId = p.EntityId
	subprocess.SigningType = p.SigningType
	subprocess.SignerDltAccountId = p.SignerDltAccountId
	p.Subprocesses = append(p.Subprocesses, subprocess)
	return true
}

func (p *Process) setSubprocessStatus(processId uuid.UUID, status ProcessStatus, revertReason string) bool {
	index := p.indexOfSubprocess(processId)
	if index < 0 {
		return false
	}
	p.Subprocesses[index].Status = status
	if revertReason != "" {
		p.Subprocesses[index].RevertReason = revertReason
	}
	return true
}

func (p *Process) indexOfSubprocess(processId uuid.UUID) int {
	for i := range p.Subprocesses {
		if p.Subprocesses[i].Id == processId {
			return i
		}
	}
	return -1
}
