package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
)

func RegisterPipeForAllTables() error {

	tables, err := LoadAuroraTablesWithColumns(NameWithPrefix(LayerToSchema(misc.LayerBronze)))
	if err != nil {
		return err
	}

	for _, table := range tables {
		if err = RegisterPipeForTable(table.RelName.String); err != nil {
			return fmt.Errorf("failed to register pipe for table %s: %w", table.RelName.String, err)
		}
	}

	return nil

}

func RegisterPipeForTable(tableName string) error {

	var (
		err              error
		outputsRoles     map[string]string
		outputsStreaming map[string]string
	)

	cfg := GetAWSConfig()
	client := pipes.NewFromConfig(GetAWSConfig())

	outputsRoles, err = (&Stack{Name: misc.StackNameRoles}).GetStackOutputs(cfg)
	if err != nil {
		return fmt.Errorf("failed to get outputs from roles stack: %w", err)
	}

	outputsStreaming, err = (&Stack{Name: misc.StackNameStreaming}).GetStackOutputs(cfg)
	if err != nil {
		return fmt.Errorf("failed to get outputs from stack %s: %w", misc.StackNameStreaming, err)
	}

	pipeName := fmt.Sprintf("dm-pipe-%s", tableName)
	fmt.Printf("Creating pipe %s for table %s...\n", pipeName, tableName)

	_, err = client.DescribePipe(context.TODO(), &pipes.DescribePipeInput{
		Name: aws.String(pipeName),
	})
	if err == nil {
		fmt.Println(misc.Yellow("Pipe %s already exists, skipping creation", pipeName))
		return nil
	}

	var conflict *types.ConflictException
	if errors.As(err, &conflict) {
		fmt.Println(misc.Yellow("Pipe %s already exists (409), skipping", pipeName))
		return nil
	}

	_, err = client.CreatePipe(context.TODO(), &pipes.CreatePipeInput{
		Name:    aws.String(pipeName),
		RoleArn: aws.String(outputsRoles["PipeExecutionRoleArn"]),
		Source:  aws.String(outputsStreaming["KinesisStreamArn"]),
		Target:  aws.String(outputsStreaming["FirehoseStreamArn"]),
		SourceParameters: &types.PipeSourceParameters{
			KinesisStreamParameters: &types.PipeSourceKinesisStreamParameters{
				StartingPosition: types.KinesisStreamStartPositionLatest,
			},
		},
		TargetParameters: &types.PipeTargetParameters{
			InputTemplate: aws.String(`{
			  "table": "<$.metadata.table-name>",
			  "op": "<$.metadata.operation>",
			  "payload": <$.data>
			}`),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create pipe for table %s: %w", tableName, err)
	}

	fmt.Println(misc.Green("Pipe %s registered for table: %s", pipeName, tableName))
	return nil

}
