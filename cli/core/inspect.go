package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/seriallink/datamaster/cli/misc"
)

func InspectStream(tableName string) error {

	client := kinesis.NewFromConfig(GetAWSConfig())

	outputsStreaming, err := GetStackOutputs(&Stack{Name: misc.StackNameStreaming})
	if err != nil {
		return fmt.Errorf("failed to get outputs from streaming stack: %w", err)
	}

	streamArn := outputsStreaming["KinesisStreamArn"]
	if streamArn == "" {
		return fmt.Errorf("KinesisStreamArn not found in outputs")
	}

	streamName := ExtractNameFromArn(streamArn)
	desc, err := client.DescribeStream(context.TODO(), &kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		return fmt.Errorf("failed to describe stream: %w", err)
	}

	if len(desc.StreamDescription.Shards) == 0 {
		return fmt.Errorf("no shards found in stream: %s", streamName)
	}

	shardID := desc.StreamDescription.Shards[0].ShardId

	iteratorOutput, err := client.GetShardIterator(context.TODO(), &kinesis.GetShardIteratorInput{
		StreamName:        aws.String(streamName),
		ShardId:           shardID,
		ShardIteratorType: types.ShardIteratorTypeLatest,
	})
	if err != nil {
		return fmt.Errorf("failed to get shard iterator: %w", err)
	}

	iterator := iteratorOutput.ShardIterator
	time.Sleep(2 * time.Second)

	recordsOutput, err := client.GetRecords(context.TODO(), &kinesis.GetRecordsInput{
		ShardIterator: iterator,
		Limit:         aws.Int32(5),
	})
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	if len(recordsOutput.Records) == 0 {
		fmt.Println(misc.Yellow("No records found in stream."))
		return nil
	}

	fmt.Println(misc.Blue("Recent records in stream:") + "\n")
	for _, rec := range recordsOutput.Records {
		var payload map[string]interface{}
		err = json.Unmarshal(rec.Data, &payload)
		if err != nil {
			fmt.Println(misc.Red("[Invalid JSON]"), string(rec.Data))
			continue
		}

		pretty, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(pretty))
		fmt.Println("---")
	}

	return nil
}
