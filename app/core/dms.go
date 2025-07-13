package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
)

var cdcOnce sync.Once

// StartReplication initiates or resumes an AWS DMS replication task based on its current status.
// If the task is in "ready" state, it starts the task from the beginning.
// If the task is "stopped", it resumes processing. If already "running", no action is taken.
// The function waits until the task reaches the "running" state before returning.
//
// Returns:
//   - error: an error if the task cannot be described, started, resumed, or confirmed as running.
func StartReplication() error {

	client := databasemigrationservice.NewFromConfig(GetAWSConfig())

	taskArn, err := (&Stack{Name: misc.StackNameStreaming}).GetStackOutput(GetAWSConfig(), "DMSReplicationTaskArn")
	if err != nil {
		return fmt.Errorf("DMSReplicationTaskArn not found in stack outputs")
	}

	fmt.Println("CDC replication is starting — this may take several minutes on the first run...")
	fmt.Printf("You can monitor progress via AWS Console > DMS > %s\n", taskArn)

	desc, err := client.DescribeReplicationTasks(context.TODO(), &databasemigrationservice.DescribeReplicationTasksInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("replication-task-arn"),
				Values: []string{taskArn},
			},
		},
	})
	if err != nil || len(desc.ReplicationTasks) == 0 {
		return fmt.Errorf("failed to describe replication task: %w", err)
	}

	status := desc.ReplicationTasks[0].Status
	fmt.Printf("Current task status: %s\n", *status)

	switch *status {
	case "ready":
		fmt.Println("Starting task for the first time (StartReplication)...")
		_, err = client.StartReplicationTask(context.TODO(), &databasemigrationservice.StartReplicationTaskInput{
			ReplicationTaskArn:       aws.String(taskArn),
			StartReplicationTaskType: types.StartReplicationTaskTypeValueStartReplication,
		})
	case "stopped":
		fmt.Println("Resuming task (ResumeProcessing)...")
		_, err = client.StartReplicationTask(context.TODO(), &databasemigrationservice.StartReplicationTaskInput{
			ReplicationTaskArn:       aws.String(taskArn),
			StartReplicationTaskType: types.StartReplicationTaskTypeValueResumeProcessing,
		})
	case "running":
		fmt.Println("Replication task is already running.")
		return nil
	default:
		return fmt.Errorf("replication task is in unsupported state: %s", status)
	}

	if err != nil {
		return fmt.Errorf("failed to start replication task: %w", err)
	}

	return waitUntilTaskRunning(client, taskArn)

}

// waitUntilTaskRunning polls the replication task status until it reaches "running" or times out.
// It checks the task status every 6 seconds, up to a maximum of 30 attempts (~3 minutes).
//
// Parameters:
//   - client: the AWS DMS client used to query task status.
//   - taskArn: the ARN of the replication task to monitor.
//
// Returns:
//   - error: an error if the task is not found, fails to start, or does not reach "running" state within the expected time.
func waitUntilTaskRunning(client *databasemigrationservice.Client, taskArn string) error {
	for i := 0; i < 30; i++ {
		out, err := client.DescribeReplicationTasks(context.TODO(), &databasemigrationservice.DescribeReplicationTasksInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("replication-task-arn"),
					Values: []string{taskArn},
				},
			},
		})
		if err != nil {
			return err
		}
		if len(out.ReplicationTasks) == 0 {
			return fmt.Errorf("replication task not found")
		}
		status := out.ReplicationTasks[0].Status
		fmt.Printf("Task status: %s\n", *status)
		if *status == "running" {
			fmt.Println("Replication task is running.")
			return nil
		}
		time.Sleep(6 * time.Second)
	}
	return fmt.Errorf("replication task did not start within expected time")
}

// ensureCDCStarted runs the CDC replication startup logic only once using sync.Once.
// If the replication is already running, it logs a message and clears the error.
//
// Returns:
//   - error: an error if the replication task failed to start and was not already running.
func ensureCDCStarted() (err error) {
	cdcOnce.Do(func() {
		err = StartReplication()
		if isAlreadyStartedError(err) {
			fmt.Println("CDC replication already running.")
			err = nil
		}
	})
	return
}

// isAlreadyStartedError checks whether the given error indicates that the DMS replication task is already running.
//
// Parameters:
//   - err: the error to inspect.
//
// Returns:
//   - bool: true if the error message contains "AlreadyStartedFault"; false otherwise.
func isAlreadyStartedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "AlreadyStartedFault")
}
