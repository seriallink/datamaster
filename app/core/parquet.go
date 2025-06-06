package core

import (
	"fmt"
	"io"
	"reflect"

	"github.com/parquet-go/parquet-go"
)

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

	pw := parquet.NewWriter(writer, parquet.SchemaOf(v.Index(0).Interface()))
	defer pw.Close()

	for i := 0; i < v.Len(); i++ {
		if err := pw.Write(v.Index(i).Interface()); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}

	return nil

}
