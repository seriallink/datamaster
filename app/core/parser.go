package core

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/misc"
)

// UnmarshalRecords converts a slice of generic maps into a typed slice using reflection.
//
// This function uses the type of the `model` parameter to infer the target slice type,
// allocates a new slice of that type, e preenche os valores usando `misc.Copier`.
//
// Parameters:
//   - model: An instance of the target struct type (e.g., &MyStruct{}) used to determine the slice element type.
//   - data: A slice of `map[string]any`, typically obtained from JSON or dynamic decoding.
//
// Returns:
//   - any: A pointer to a slice of the inferred type, e.g., *[]MyStruct.
//   - error: An error if the conversion fails.
func UnmarshalRecords(model any, data []map[string]any) (any, error) {

	// expected return type
	rt := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(model)).Type())

	// init slice with the correct type
	slice := reflect.New(rt).Interface()

	err := misc.Copier(slice, data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal records: %w", err)
	}

	return slice, nil

}

// CsvToMap reads a CSV stream and converts each valid row into a map[string]any.
//
// This function parses the CSV header to determine column names and processes each row,
// trimming whitespace and converting empty strings to `nil`. Each row is validated using
// the provided model via `ValidateRow`. Invalid rows are skipped with warnings.
//
// Parameters:
//   - r: An `io.Reader` containing the CSV data.
//   - model: A reference model used to validate each row's structure and content.
//
// Returns:
//   - []map[string]any: A slice of validated row maps with column names as keys.
//   - error: An error if reading the header fails.
func CsvToMap(r io.Reader, model any) ([]map[string]any, error) {

	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = false

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var valid []map[string]any
	lineNumber := 1 // header line already read

	for {
		record, err := reader.Read()
		lineNumber++

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("CSV parse error on line %d: %v", lineNumber, err)
			continue
		}
		if len(record) != len(headers) {
			log.Printf("CSV malformed row at line %d: expected %d fields, got %d", lineNumber, len(headers), len(record))
			continue
		}

		row := make(map[string]any, len(headers))
		for i, col := range headers {
			val := strings.TrimSpace(record[i])
			if val == "" {
				row[col] = nil
			} else {
				row[col] = val
			}
		}

		if err = ValidateRow(row, model, lineNumber); err != nil {
			log.Printf("Row rejected on line %d: %v", lineNumber, err)
			// TODO: save to /purge if needed
			continue
		}

		valid = append(valid, row)
	}

	return valid, nil

}

// ValidateRow checks if a given CSV row map conforms to the structure and types defined in a model struct.
//
// It verifies the presence of all required fields (based on JSON tags) and attempts to convert values
// to the expected types. Pointer fields are allowed to be nil. Supported types include string, int64,
// float64, json.Number, and time.Time (RFC3339). Unsupported types result in an error.
//
// Parameters:
//   - row: A map representing a parsed CSV row, with column names as keys.
//   - model: A struct type used to validate field presence and types (must be a pointer to struct).
//   - line: The line number in the CSV file, used for error context.
//
// Returns:
//   - error: A validation error if the row is invalid, otherwise nil.
func ValidateRow(row map[string]any, model any, line int) error {

	t := reflect.TypeOf(model).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		colName := strings.Split(jsonTag, ",")[0]
		val, exists := row[colName]
		if !exists {
			return fmt.Errorf("line %d: missing field: %s", line, colName)
		}

		// allow nil for pointer fields
		if val == nil {
			if field.Type.Kind() == reflect.Ptr {
				continue // valid: nil pointer
			}
			return fmt.Errorf("line %d: missing required value in column %s", line, colName)
		}

		// normalize type (handle pointers)
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		strVal := fmt.Sprintf("%v", val)

		switch field.Type.String() {
		case "int64", "*int64":
			if _, err := strconv.ParseInt(strVal, 10, 64); err != nil {
				return fmt.Errorf("line %d: invalid int for column %s: %v", line, colName, strVal)
			}
		case "float64", "*float64":
			if _, err := strconv.ParseFloat(strVal, 64); err != nil {
				return fmt.Errorf("line %d: invalid float for column %s: %v", line, colName, strVal)
			}
		case "json.Number", "*json.Number":
			if _, err := strconv.ParseFloat(strVal, 64); err != nil {
				return fmt.Errorf("line %d: invalid json.Number for column %s: %v", line, colName, strVal)
			}
		case "time.Time", "*time.Time":
			if _, err := time.Parse(time.RFC3339, strVal); err != nil {
				return fmt.Errorf("line %d: invalid timestamp for column %s: %v", line, colName, strVal)
			}
		case "string", "*string":
			// always valid
		default:
			// fallback using kind
			switch ft.Kind() {
			case reflect.Int, reflect.Int64:
				if _, err := strconv.ParseInt(strVal, 10, 64); err != nil {
					return fmt.Errorf("line %d: invalid int for column %s: %v", line, colName, strVal)
				}
			case reflect.Float64:
				if _, err := strconv.ParseFloat(strVal, 64); err != nil {
					return fmt.Errorf("line %d: invalid float for column %s: %v", line, colName, strVal)
				}
			case reflect.String:
				// ok
			default:
				return fmt.Errorf("line %d: unsupported type for column %s: %s", line, colName, field.Type.String())
			}
		}
	}

	return nil

}
