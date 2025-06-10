package vmstate

type VMState int

const (
	STOPPED   VMState = iota // STOPPED is the state when the VM is not running
	OFFLINE                  // OFFLINE is the state when the VM is running but not reachable
	ONLINE                   // ONLINE is the state when the VM is running and reachable
	RUNNING                  // RUNNING is the state when the VM is running and executing commands
	RESTORING                // RESTORING is the state when the VM is restoring a snapshot
	ERROR                    // ERROR is the state when the VM is in an error state
)

func (s VMState) String() string {
	switch s {
	case STOPPED:
		return "STOPPED"
	case OFFLINE:
		return "OFFLINE"
	case ONLINE:
		return "ONLINE"
	case RUNNING:
		return "RUNNING"
	case RESTORING:
		return "RESTORING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
