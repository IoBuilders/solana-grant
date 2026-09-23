package filter

// SyncStatus describes where a Filter stands in catching up with its Node's
// chain history: actively caught up (ACTIVE), replaying historical slots
// (SYNCING) or paused with no work pending (IDLE).
type SyncStatus string

const (
	SyncStatusActive  SyncStatus = "ACTIVE"
	SyncStatusSyncing SyncStatus = "SYNCING"
	SyncStatusIdle    SyncStatus = "IDLE"
)

func (s SyncStatus) IsValid() bool {
	switch s {
	case SyncStatusActive, SyncStatusSyncing, SyncStatusIdle:
		return true
	}
	return false
}

func (s SyncStatus) String() string {
	return string(s)
}
