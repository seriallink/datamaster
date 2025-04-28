package core

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/abiosoft/ishell"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/fatih/color"
)

func DeployAllStacks(c *ishell.Context, templates embed.FS) error {

	// deploy network
	c.Println(misc.Blue("Deploying network stack..."))
	networkContent, _ := templates.ReadFile("infra/templates/network.yml")
	if err := DeployStack("dm-network", networkContent, nil); err != nil {
		return fmt.Errorf("failed to deploy network stack: %w", err)
	}

	// get vpc and subnet ids
	networkOutput, err := GetStackOutputs("dm-network")
	if err != nil {
		return fmt.Errorf("failed to get network stack outputs: %w", err)
	}

	// deploy security
	c.Println(misc.Blue("Deploying security stack..."))
	securityContent, _ := templates.ReadFile("infra/templates/security.yml")
	securityParams := []types.Parameter{
		{ParameterKey: aws.String("VpcId"), ParameterValue: aws.String(networkOutput["VpcId"])},
	}
	if err = DeployStack("dm-security", securityContent, securityParams); err != nil {
		return fmt.Errorf("failed to deploy security stack: %w", err)
	}

	// get sg id and secret arn
	securityOutput, err := GetStackOutputs("dm-security")

	// deploy aurora
	c.Println(misc.Blue("Deploying aurora stack..."))
	auroraContent, _ := templates.ReadFile("infra/templates/aurora.yml")
	auroraParams := []types.Parameter{
		{ParameterKey: aws.String("VpcId"), ParameterValue: aws.String(networkOutput["VpcId"])},
		{ParameterKey: aws.String("SubnetId"), ParameterValue: aws.String(networkOutput["VpcId"])},
		{ParameterKey: aws.String("SecurityGroupId"), ParameterValue: aws.String(securityOutput["SecurityGroupId"])},
		{ParameterKey: aws.String("SecretArn"), ParameterValue: aws.String(securityOutput["SecretArn"])},
	}
	if err = DeployStack("dm-aurora", auroraContent, auroraParams); err != nil {
		return fmt.Errorf("failed to deploy aurora stack: %w", err)
	}

	return nil

}

func DeployStack(stackName string, templateContent []byte, parameters []types.Parameter) error {
	cfClient := cloudformation.NewFromConfig(awsConfig)

	// first attempt to create the stack
	input := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(string(templateContent)),
		Parameters:   parameters,
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
			types.CapabilityCapabilityNamedIam,
		},
	}

	_, err := cfClient.CreateStack(context.TODO(), input)
	if err != nil {
		if strings.Contains(err.Error(), "AlreadyExists") {
			color.Yellow("Stack %s already exists. Attempting to update...", stackName)

			updateInput := &cloudformation.UpdateStackInput{
				StackName:    aws.String(stackName),
				TemplateBody: aws.String(string(templateContent)),
				Parameters:   parameters,
				Capabilities: []types.Capability{
					types.CapabilityCapabilityIam,
					types.CapabilityCapabilityNamedIam,
				},
			}

			_, err = cfClient.UpdateStack(context.TODO(), updateInput)
			if err != nil {
				if strings.Contains(err.Error(), "No updates are to be performed") {
					color.Green("No updates needed for stack %s", stackName)
					return nil
				}
				return fmt.Errorf("failed to update stack %s: %w", stackName, err)
			}
		} else {
			return fmt.Errorf("failed to create stack %s: %w", stackName, err)
		}
	}

	color.Green("Waiting for stack %s to complete...", stackName)

	// wait for the stack to be created or updated
	waiter := cloudformation.NewStackCreateCompleteWaiter(cfClient)
	err = waiter.Wait(context.TODO(), &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("stack %s did not reach complete state: %w", stackName, err)
	}

	color.Green("Stack %s completed successfully", stackName)
	return nil

}

func GetStackOutputs(stackName string) (map[string]string, error) {

	cfClient := cloudformation.NewFromConfig(awsConfig)

	resp, err := cfClient.DescribeStacks(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe stack %s: %w", stackName, err)
	}

	if len(resp.Stacks) == 0 {
		return nil, fmt.Errorf("no stacks found with name %s", stackName)
	}

	outputs := make(map[string]string)
	for _, o := range resp.Stacks[0].Outputs {
		if o.OutputKey != nil && o.OutputValue != nil {
			outputs[*o.OutputKey] = *o.OutputValue
		}
	}

	return outputs, nil

}
