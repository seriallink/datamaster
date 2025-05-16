package misc

import (
	"time"

	"github.com/google/uuid"
)

func TernaryG[T any](condition bool, x, y T) T {
	if condition {
		return x
	}
	return y
}

func Ternary(condition bool, x, y any) any {
	if condition {
		return x
	}
	return y
}

func TernaryStr(condition bool, x, y string) string {
	return Ternary(condition, x, y).(string)
}

func TernaryInt(condition bool, x, y int) int {
	return Ternary(condition, x, y).(int)
}

func TernaryFloat(condition bool, x, y float64) float64 {
	return Ternary(condition, x, y).(float64)
}

func TernaryUUID(condition bool, x, y uuid.UUID) uuid.UUID {
	return Ternary(condition, x, y).(uuid.UUID)
}

func TernaryErr(condition bool, x, y error) error {
	return Ternary(condition, x, y).(error)
}

func TernaryTime(condition bool, x, y time.Time) time.Time {
	return Ternary(condition, x, y).(time.Time)
}
