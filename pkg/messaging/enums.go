package messaging

type LoginFailedReason uint8

const (
	LoginFailed      LoginFailedReason = 6
	UpdateAvailable                    = 8
	ConnectionError                    = 9
	MaintenanceBreak                   = 10
	Banned                             = 11
	AccountLocked                      = 13
)
