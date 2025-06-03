package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
)

var cdcOnce sync.Once

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
		fmt.Printf("⌛ Task status: %s\n", *status)
		if *status == "running" {
			fmt.Println("Replication task is running.")
			return nil
		}
		time.Sleep(6 * time.Second)
	}
	return fmt.Errorf("replication task did not start within expected time")
}

func ensureCDCStarted() error {
	var startErr error

	cdcOnce.Do(func() {
		startErr = StartReplication()
		if isAlreadyStartedError(startErr) {
			fmt.Println("CDC replication already running.")
			startErr = nil // limpa o erro para não bloquear
		}
	})

	return startErr
}

func isAlreadyStartedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "AlreadyStartedFault")
}
