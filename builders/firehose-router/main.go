package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Payload struct {
	Data     json.RawMessage `json:"data"`
	Metadata struct {
		TableName  string `json:"table-name"`
		Operation  string `json:"operation"`
		RecordType string `json:"record-type"`
		SchemaName string `json:"schema-name"`
		Timestamp  string `json:"timestamp"`
	} `json:"metadata"`
}

func handler(ctx context.Context, event events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {

	var response events.KinesisFirehoseResponse

	accepted := make(map[string]int)
	dropped := make(map[string]int)

	log.Printf("Received %d record(s)", len(event.Records))

	for _, record := range event.Records {
		var payload Payload
		if err := json.Unmarshal(record.Data, &payload); err != nil {
			dropWithLog(&response, record.RecordID, "unparsed", "failed to unmarshal record.Data", dropped)
			continue
		}

		table := payload.Metadata.TableName
		if table == "" {
			dropWithLog(&response, record.RecordID, "unknown", "missing table-name in metadata", dropped)
			continue
		}

		var flat map[string]interface{}
		if err := json.Unmarshal(payload.Data, &flat); err != nil {
			dropWithLog(&response, record.RecordID, table, "failed to flatten payload.Data", dropped)
			continue
		}
		flat["operation"] = payload.Metadata.Operation

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(flat); err != nil {
			dropWithLog(&response, record.RecordID, table, "failed to marshal final JSON", dropped)
			continue
		}

		response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
			RecordID: record.RecordID,
			Result:   events.KinesisFirehoseTransformedStateOk,
			Data:     buf.Bytes(),
			Metadata: events.KinesisFirehoseResponseRecordMetadata{
				PartitionKeys: map[string]string{
					"table_name": table,
				},
			},
		})

		accepted[table]++

	}

	log.Println("=== Processing summary ===")
	if len(accepted) == 0 && len(dropped) == 0 {
		log.Println("No records processed.")
	} else {
		log.Println("Accepted:")
		for k, v := range accepted {
			log.Printf("  - %s: %d", k, v)
		}
		log.Println("Dropped:")
		for k, v := range dropped {
			log.Printf("  - %s: %d", k, v)
		}
	}

	return response, nil

}

func dropWithLog(response *events.KinesisFirehoseResponse, recordID, table string, reason string, dropped map[string]int) {
	log.Printf("[DROPPED] table=%s - reason=%s", table, reason)
	dropped[table]++
	response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
		RecordID: recordID,
		Result:   events.KinesisFirehoseTransformedStateDropped,
	})
}

func main() {
	lambda.Start(handler)
}
