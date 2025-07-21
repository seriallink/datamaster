package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

// TriggerProcessing starts a Step Function execution for the specified layer and optional tables.
//
// Parameters:
//   - cfg: AWS configuration used to connect to the Step Functions service.
//   - layer: processing layer (e.g. "silver", "gold").
//   - tables: optional list of table names.
//
// Returns:
//   - error: any error that occurs while triggering the pipeline.
func TriggerProcessing(cfg aws.Config, layer string, tables []string) error {

	var (
		err error
		arn string
	)

	switch layer {
	case misc.LayerSilver:
		arn, err = (&Stack{Name: misc.StackNameProcessing}).GetStackOutput(cfg, "ProcessingDispatcherArn")
	case misc.LayerGold:
		arn, err = (&Stack{Name: misc.StackNameAnalytics}).GetStackOutput(cfg, "AnalyticsDispatcherArn")
	default:
		return fmt.Errorf("invalid layer specified: %s. Valid options are: %s, %s", layer, misc.LayerSilver, misc.LayerGold)
	}
	if err != nil {
		return fmt.Errorf("error retrieving dispatcher ARN: %w", err)
	}

	if len(tables) == 0 {
		tables, err = FetchPendingTables(cfg, NameWithPrefix(layer))
		if err != nil {
			return err
		}
		if len(tables) == 0 {
			fmt.Println(misc.Yellow("No tables to process in layer '%s'. Skipping execution.", layer))
			return nil
		}
	}

	input := map[string]interface{}{
		"tables": tables,
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error marshalling Step Function input: %w", err)
	}

	client := sfn.NewFromConfig(cfg)

	execName := fmt.Sprintf("manual-%s", time.Now().Format("20060102-150405"))

	_, err = client.StartExecution(context.TODO(), &sfn.StartExecutionInput{
		StateMachineArn: aws.String(arn),
		Input:           aws.String(string(payload)),
		Name:            aws.String(execName),
	})
	if err != nil {
		return fmt.Errorf("error starting Step Function execution: %w", err)
	}

	return nil

}
