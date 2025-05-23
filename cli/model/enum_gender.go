package model

type GenderEnum string

const (
	GenderMasculine GenderEnum = "masculine"
	GenderFeminine  GenderEnum = "feminine"
)

func (enum GenderEnum) Values() []Enum {
	return []Enum{
		GenderMasculine,
		GenderFeminine,
	}
}

func (enum GenderEnum) String() string {
	return string(enum)
}

func (enum GenderEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum GenderEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
