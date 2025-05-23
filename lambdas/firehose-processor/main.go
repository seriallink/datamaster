package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {

	var response events.KinesisFirehoseResponse

	log.Printf("Received %d record(s)", len(event.Records))

	for _, record := range event.Records {

		var recordMap map[string]interface{}
		if err := json.Unmarshal(record.Data, &recordMap); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateDropped,
			})
			continue
		}

		var tableName string
		if payloadField, ok := recordMap["payload"].(map[string]interface{}); ok {
			if metadata, ok := payloadField["metadata"].(map[string]interface{}); ok {
				if name, ok := metadata["table-name"].(string); ok && name != "" {
					tableName = name
				}
			}
		}
		if tableName == "" {
			tableName = "unknown"
		}

		response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
			RecordID: record.RecordID,
			Result:   events.KinesisFirehoseTransformedStateOk,
			Data:     record.Data,
			Metadata: events.KinesisFirehoseResponseRecordMetadata{
				PartitionKeys: map[string]string{
					"table_name": tableName,
				},
			},
		})
	}

	return response, nil

}

func main() {
	lambda.Start(handler)
}
