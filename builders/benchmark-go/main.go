package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/seriallink/datamaster/app/bronze"
	"github.com/seriallink/datamaster/app/core"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/parquet-go/parquet-go"
)

type job struct {
	Index int
	Row   []string
}

func main() {

	start := time.Now()
	ctx := context.Background()

	bucket := os.Getenv("BENCHMARK_BUCKET")
	input := os.Getenv("BENCHMARK_INPUT")
	output := os.Getenv("BENCHMARK_OUTPUT")
	result := os.Getenv("BENCHMARK_RESULT")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	s3client := s3.NewFromConfig(cfg)

	resp, err := s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(input),
	})
	if err != nil {
		log.Fatalf("s3.GetObject error: %v", err)
	}
	defer resp.Body.Close()

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("gzip.NewReader error: %v", err)
	}
	defer gzr.Close()

	reader := csv.NewReader(gzr)
	reader.ReuseRecord = true
	headers, err := reader.Read()
	if err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}

	// Index map
	idx := make(map[string]int)
	for i, name := range headers {
		idx[name] = i
	}

	// Channels
	jobs := make(chan job, 10000)
	parsed := make(chan *bronze.Review, 10000)

	// Worker pool: parse CSV to struct
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU() * 2
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				obj, err := new(bronze.Review).FromCSV(j.Row, idx)
				if err == nil {
					parsed <- obj.(*bronze.Review)
				}
			}
		}()
	}

	// Read and dispatch jobs
	go func() {
		i := 0
		for {
			row, err := reader.Read()
			if err != nil {
				break
			}
			jobs <- job{Index: i, Row: append([]string(nil), row...)}
			i++
		}
		close(jobs)
	}()

	// Close parsed after workers finish
	go func() {
		wg.Wait()
		close(parsed)
	}()
	csvElapsed := time.Since(start)
	log.Printf("CSV read + parse done in %s", csvElapsed)

	// Parquet writer with batching
	parquetStart := time.Now()
	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[*bronze.Review](&buf,
		parquet.Compression(&parquet.Snappy),
		parquet.DefaultEncoding(&parquet.Plain),
	)

	const batchSize = 1000
	batch := make([]*bronze.Review, 0, batchSize)

	for rec := range parsed {
		batch = append(batch, rec)
		if len(batch) >= batchSize {
			if _, err := writer.Write(batch); err != nil {
				log.Fatalf("writer.Write error: %v", err)
			}
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		if _, err := writer.Write(batch); err != nil {
			log.Fatalf("writer.Write error: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		log.Fatalf("writer.Close error: %v", err)
	}
	parquetElapsed := time.Since(parquetStart)
	log.Printf("Parquet write done in %s", parquetElapsed)
	log.Printf("Parquet size: %.2f MB", float64(buf.Len())/(1024*1024))

	// Upload to S3
	_, err = s3client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(output),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		log.Fatalf("PutObject error: %v", err)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	totalElapsed := time.Since(start)
	memUsedMB := float64(mem.Alloc) / (1024 * 1024)

	log.Printf("Total time: %s", totalElapsed)
	log.Printf("Memory used: %.2f MB", memUsedMB)

	// Create .json result
	summary := core.BenchmarkResult{
		Implementation:   "go",
		StartTaskTime:    start,
		EndTaskTime:      time.Now(),
		InputFile:        input,
		OutputFile:       output,
		CSVReadTime:      csvElapsed,
		ParquetWriteTime: parquetElapsed,
		TotalTaskTime:    totalElapsed,
		MemoryUsedMB:     memUsedMB,
	}
	jsonBuf, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Fatalf("MarshalIndent error: %v", err)
	}

	_, err = s3client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(result),
		Body:   bytes.NewReader(jsonBuf),
	})
	if err != nil {
		log.Fatalf("PutObject (json) error: %v", err)
	}

}
