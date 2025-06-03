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

	log.Printf("Received %d record(s)", len(event.Records))

	for _, record := range event.Records {

		var payload Payload
		if err := json.Unmarshal(record.Data, &payload); err != nil {
			log.Printf("Failed to parse raw DMS event: %v", err)
			response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateDropped,
			})
			continue
		}

		if payload.Metadata.TableName == "" {

			payload.Metadata.TableName = "unknown"

			outputMap := map[string]interface{}{
				"table":   payload.Metadata.TableName,
				"op":      payload.Metadata.Operation,
				"payload": payload.Data,
			}

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(outputMap); err != nil {
				log.Printf("Failed to marshal fallback output JSON: %v", err)
				response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
					RecordID: record.RecordID,
					Result:   events.KinesisFirehoseTransformedStateDropped,
				})
				continue
			}

			response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateOk,
				Data:     buf.Bytes(),
				Metadata: events.KinesisFirehoseResponseRecordMetadata{
					PartitionKeys: map[string]string{
						"table_name": "unknown",
					},
				},
			})

			continue // skip flattening

		}

		// Flatten payload and attach operation field
		var flat map[string]interface{}
		if err := json.Unmarshal(payload.Data, &flat); err != nil {
			log.Printf("Failed to parse payload.Data into flat map: %v", err)
			response.Records = append(response.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateDropped,
			})
			continue
		}
		flat["operation"] = payload.Metadata.Operation

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(flat); err != nil {
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
			Data:     buf.Bytes(),
			Metadata: events.KinesisFirehoseResponseRecordMetadata{
				PartitionKeys: map[string]string{
					"table_name": payload.Metadata.TableName,
				},
			},
		})

	}

	return response, nil

}

func main() {
	lambda.Start(handler)
}
