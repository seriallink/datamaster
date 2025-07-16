package misc

import (
	"database/sql"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/metakeule/fmtdate"
	"github.com/seriallink/null"
)

// Trim removes leading/trailing whitespace from the input string,
// after first removing any common indentation using Dedent.
//
// It is useful for formatting multi-line string literals where indentation
// is used for code readability but should not be preserved.
//
// Example:
//
//	input := `
//	    SELECT *
//	    FROM users
//	`
//	cleaned := Trim(input)
//
// Returns:
//   - string: trimmed and dedented version of the input.
func Trim(s string) string {
	return strings.TrimSpace(Dedent(s))
}

// Dedent removes common leading whitespace from every line of the input string.
// This is typically used for multi-line strings defined inline in code.
//
// It finds the smallest indentation (spaces or tabs) among all non-empty lines
// and removes that indentation from every line.
//
// Example:
//
//	input := `
//	    line 1
//	    line 2
//	`
//	output := Dedent(input)
//
// Returns:
//   - string: dedented version of the input.
func Dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1
	for _, line := range lines {
		if trimmed := strings.TrimLeft(line, " \t"); trimmed != "" {
			indent := len(line) - len(trimmed)
			if minIndent == -1 || indent < minIndent {
				minIndent = indent
			}
		}
	}
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}
	return strings.Join(lines, "\n")
}

// ToPascalCase converts a kebab-case string into PascalCase.
//
// Parameters:
//   - input: string - The input string in kebab-case format (e.g., "my-variable-name").
//
// Returns:
//   - string: The resulting string in PascalCase format (e.g., "MyVariableName").
func ToPascalCase(input string) string {
	parts := strings.Split(input, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// AnyToString converts a generic value to its string representation.
//
// Parameters:
//   - value: a value of any supported type.
//
// Returns:
//   - string: the string representation of the input value.
//
// Supported types:
//   - Integer types (int, int8, int16, int32, int64)
//   - Unsigned integers (uint, uint8, uint16, uint32, uint64)
//   - Floating-point numbers (float32, float64)
//   - Booleans
//   - Strings and []byte
//   - []string (joined with commas)
//   - Nullable types (null.String, sql.NullString, null.Time)
//   - time.Time
//   - UUIDs (uuid.UUID and *uuid.UUID)
//   - nil (returns empty string)
//   - Other types are converted using Stringify (fallback).
func AnyToString(value any) string {

	switch v := value.(type) {

	case int:
		return strconv.Itoa(v)
	case int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)

	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)

	case bool:
		return strconv.FormatBool(v)

	case []string:
		return strings.Join(v, ",")
	case []byte:
		return string(v)

	case string:
		return v

	case null.String:
		return v.String
	case sql.NullString:
		return v.String

	case time.Time:
		return fmtdate.FormatDateTime(v)
	case null.Time:
		if v.Valid {
			return fmtdate.FormatDateTime(v.Time)
		}
		return ""

	case uuid.UUID:
		return v.String()
	case *uuid.UUID:
		if v != nil {
			return v.String()
		}
		return ""

	case nil:
		return ""

	default:
		return Stringify(v)
	}

}
