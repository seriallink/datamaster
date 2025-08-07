// Package enum centralizes all enumerated types used throughout the Data Master project.
//
// It provides strongly-typed constants for process statuses, file formats,
// and compute targets, ensuring consistency across the ingestion pipeline,
// control mechanisms, and processing components.
//
// These enums are used for controlling flow logic, determining orchestration
// strategies (e.g., Lambda vs ECS), and interpreting file content during
// transformation steps.
package enum

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
