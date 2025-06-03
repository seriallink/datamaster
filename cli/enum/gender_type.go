package enum

type GenderTypeEnum string

const (
	GenderMasculine GenderTypeEnum = "masculine"
	GenderFeminine  GenderTypeEnum = "feminine"
)

func (enum GenderTypeEnum) Values() []Enum {
	return []Enum{
		GenderMasculine,
		GenderFeminine,
	}
}

func (enum GenderTypeEnum) String() string {
	return string(enum)
}

func (enum GenderTypeEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum GenderTypeEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
