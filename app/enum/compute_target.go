package enum

type ComputeTargetEnum string

const (
	ComputeTargetLambda ComputeTargetEnum = "lambda"
	ComputeTargetECS    ComputeTargetEnum = "ecs"
	ComputeTargetEMR    ComputeTargetEnum = "emr"
)

func (enum ComputeTargetEnum) Values() []Enum {
	return []Enum{
		ComputeTargetLambda,
		ComputeTargetECS,
		ComputeTargetEMR,
	}
}

func (enum ComputeTargetEnum) String() string {
	return string(enum)
}

func (enum ComputeTargetEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum ComputeTargetEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
