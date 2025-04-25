package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/fatih/color"
)

//func DeployAllStacks(stacks embed.FS) error {
//	// Lista de stacks e templates a serem deployados
//	stacksToDeploy := []struct {
//		name     string
//		template string
//	}{
//		{"network", "network.yml"},
//		{"s3-buckets", "s3-buckets.yml"},
//		{"glue-jobs", "glue-jobs.yml"},
//		{"lakeformation", "lakeformation.yml"},
//		{"redshift", "redshift.yml"},
//	}
//
//	for _, stack := range stacksToDeploy {
//		if err := DeployStack(stack.name, stack.template); err != nil {
//			return fmt.Errorf("error deploying stack %s: %w", stack.name, err)
//		}
//	}
//
//	return nil
//}

func DeployStack(stackName string, templateContent []byte) (err error) {

	cfClient := cloudformation.NewFromConfig(awsConfig)

	input := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(string(templateContent)),
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam, // Se o template usar IAM
			types.CapabilityCapabilityNamedIam,
		},
	}

	// Cria ou atualiza a stack
	_, err = cfClient.CreateStack(context.TODO(), input)
	if err != nil {
		// Caso o stack já exista, tentamos atualizar
		if strings.Contains(err.Error(), "AlreadyExists") {
			// Atualiza o stack
			updateInput := &cloudformation.UpdateStackInput{
				StackName:    aws.String(stackName),
				TemplateBody: aws.String(string(templateContent)),
			}

			_, err = cfClient.UpdateStack(context.TODO(), updateInput)
			if err != nil {
				return fmt.Errorf("failed to update stack %s: %w", stackName, err)
			}
			color.Green("Stack updated successfully: %s", stackName)
		} else {
			return fmt.Errorf("failed to create stack %s: %w", stackName, err)
		}
	} else {
		color.Green("Stack created successfully: %s", stackName)
	}

	return nil
}
