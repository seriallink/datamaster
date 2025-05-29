package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RawDmsPayload struct {
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

	log.Printf("Received %d record(s)", len(event.Records))

	for _, record := range event.Records {

		var raw RawDmsPayload
		if err := json.Unmarshal(record.Data, &raw); err != nil {
			log.Printf("Failed to parse raw DMS event: %v", err)
			response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateDropped,
			})
			continue
		}

		if raw.Metadata.TableName == "" {
			raw.Metadata.TableName = "unknown"
		}

		outputMap := map[string]interface{}{
			"table":   raw.Metadata.TableName,
			"op":      raw.Metadata.Operation,
			"payload": raw.Data,
		}

		outputJSON, err := json.Marshal(outputMap)
		if err != nil {
			log.Printf("Failed to marshal output JSON: %v", err)
			response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateDropped,
			})
			continue
		}

		response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
			RecordID: record.RecordID,
			Result:   events.KinesisFirehoseTransformedStateOk,
			Data:     outputJSON,
			Metadata: events.KinesisFirehoseResponseRecordMetadata{
				PartitionKeys: map[string]string{
					"table_name": raw.Metadata.TableName,
				},
			},
		})
	}

	return response, nil

}

func main() {
	lambda.Start(handler)
}
