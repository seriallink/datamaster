package enum

type ProcessStatusEnum string

const (
	ProcessPending ProcessStatusEnum = "pending"
	ProcessRunning ProcessStatusEnum = "running"
	ProcessSuccess ProcessStatusEnum = "success"
	ProcessError   ProcessStatusEnum = "error"
)

func (enum ProcessStatusEnum) Values() []Enum {
	return []Enum{
		ProcessPending,
		ProcessRunning,
		ProcessSuccess,
		ProcessError,
	}
}

func (enum ProcessStatusEnum) String() string {
	return string(enum)
}

func (enum ProcessStatusEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum ProcessStatusEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
