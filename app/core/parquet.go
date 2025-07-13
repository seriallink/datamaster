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

// WriteParquet writes a slice of records to the provided io.Writer in Parquet format.
//
// The function expects a slice or pointer to slice as input. It uses the schema of the first
// element in the slice to construct the Parquet schema, compresses with Snappy, and encodes
// with Plain encoding. The function uses concurrency to prepare the records for writing.
//
// Parameters:
//   - records: A slice or pointer to a slice of structs representing the data.
//   - writer: Destination writer to which the Parquet data will be written.
//
// Returns:
//   - error: An error if writing fails, the input is invalid, or the slice is empty.
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
	pw := parquet.NewWriter(
		writer,
		schema,
		parquet.Compression(&parquet.Snappy),
		parquet.DefaultEncoding(&parquet.Plain),
	)
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

// StreamCsvToParquet reads CSV data from an input stream, parses it into model instances,
// and encodes the result into a Parquet-formatted buffer using concurrent workers.
//
// This function uses the CSV header to map column names to indices and processes each
// record concurrently. Each line is converted into a model instance via `FromCSV`.
// The resulting records are written to a Parquet buffer using Snappy compression.
//
// Parameters:
//   - r: CSV input stream (typically a .csv.gz file decompressed).
//   - model: An implementation of the bronze.Model interface used to instantiate records.
//
// Returns:
//   - *bytes.Buffer: A buffer containing the resulting Parquet data.
//   - error: An error if any step in the parsing or writing process fails.
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
