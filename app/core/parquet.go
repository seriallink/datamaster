package core

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"

	"github.com/seriallink/datamaster/app/bronze"

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

	schema := parquet.SchemaOf(v.Index(0).Interface())
	pw := parquet.NewWriter(writer, schema)
	defer pw.Close()

	type job struct {
		Index int
		Value reflect.Value
	}

	type result struct {
		Index int
		Value any
		Err   error
	}

	jobs := make(chan job, 10000)
	results := make(chan result, v.Len())

	var wg sync.WaitGroup
	workers := runtime.NumCPU() * 2

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results <- result{
					Index: job.Index,
					Value: job.Value.Interface(),
				}
			}
		}()
	}

	go func() {
		for i := 0; i < v.Len(); i++ {
			jobs <- job{Index: i, Value: v.Index(i)}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	buffer := make([]any, v.Len())
	for res := range results {
		if res.Err != nil {
			return fmt.Errorf("marshal failed at index %d: %w", res.Index, res.Err)
		}
		buffer[res.Index] = res.Value
	}

	for _, val := range buffer {
		if err := pw.Write(val); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil

}

func derefValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func StreamCsvToParquet(r io.Reader, model bronze.Model) (*bytes.Buffer, error) {

	csvReader := csv.NewReader(r)
	csvReader.ReuseRecord = true // Enable record reuse for performance, but copy each record before dispatch

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column names to their indices
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[col] = i
	}

	lines := make(chan []string, 10000)
	results := make(chan any, 10000)

	var wg sync.WaitGroup

	// Start worker goroutines to convert CSV lines into model instances
	for i := 0; i < runtime.NumCPU()*2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range lines {
				record, err := model.FromCSV(line, columnIndex)
				if err == nil {
					results <- record
				}
			}
		}()
	}

	// Producer: read CSV and send a copied version of each record to avoid overwriting
	go func() {
		for {
			record, err := csvReader.Read()
			if err != nil {
				close(lines)
				break
			}
			copied := make([]string, len(record))
			copy(copied, record)
			lines <- copied
		}
	}()

	// Closer: wait for all workers and then close the results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all model instances into a slice
	var records []any
	for rec := range results {
		records = append(records, rec)
	}

	// Write the result slice to Parquet
	buf := &bytes.Buffer{}
	if err = WriteParquet(records, buf); err != nil {
		return nil, fmt.Errorf("failed to write parquet: %w", err)
	}

	return buf, nil

}

func StreamJsonToParquet(r io.Reader, model bronze.Model) (*bytes.Buffer, error) {
	return nil, fmt.Errorf("StreamJsonToParquet not implemented")
}
