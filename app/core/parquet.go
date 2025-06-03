package core

import (
	"fmt"
	"io"
	"reflect"

	"github.com/parquet-go/parquet-go"
)

// WriteParquet writes a slice of struct records to the given writer using parquet-go.
//func WriteParquet(records any, writer io.Writer) error {
//	pw := parquet.NewWriter(writer)
//	defer pw.Close()
//	return pw.Write(records)
//}

func WriteParquet(records any, writer io.Writer) error {

	v := reflect.ValueOf(records)

	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, got %T", records)
	}

	if v.Len() == 0 {
		return fmt.Errorf("cannot write empty slice")
	}

	first := v.Index(0).Interface()
	schema := parquet.SchemaOf(first)

	pw := parquet.NewWriter(writer, schema)
	defer pw.Close()

	for i := 0; i < v.Len(); i++ {
		rec := v.Index(i).Interface()
		if err := pw.Write(rec); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}

	return nil

}
