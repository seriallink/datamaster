package model

import "github.com/jaswdr/faker"

type Enum interface {
	Values() []Enum
	String() string
	IsValid() bool
	IsEmpty() bool
}

func isEmptyEnum(enum Enum) bool {
	return enum.String() == ""
}

func isValidEnum(enum Enum, value Enum) bool {
	for _, v := range enum.Values() {
		if v == value {
			return true
		}
	}
	return false
}

func RandomEnum[T Enum](f faker.Faker, list []Enum) T {
	return list[f.IntBetween(0, len(list)-1)].(T)
}
