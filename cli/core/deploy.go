package core

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Stack struct {
	Name   string
	Prefix string
	Params []types.Parameter
}

func (s *Stack) FullStackName() string {
	prefix := s.Prefix
	if prefix == "" {
		prefix = misc.DefaultStackPrefix
	}
	return fmt.Sprintf("%s-%s", prefix, s.Name)
}

func (s *Stack) TemplateFilePath() string {
	return misc.TemplatesPath + "/" + s.Name + misc.TemplateExtension
}

func (s *Stack) GetTemplateBody(templates embed.FS) ([]byte, error) {
	templatePath := s.TemplateFilePath()
	templateBody, err := templates.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}
	return templateBody, nil
}

func DeployAllStacks(templates embed.FS) error {

	stacks := []Stack{
		{Name: misc.StackNameNetwork},
		{Name: misc.StackNameRoles},
		{Name: misc.StackNameSecurity},
		{Name: misc.StackNameStorage},
		{Name: misc.StackNameDatabase},
		{Name: misc.StackNameCatalog},
		{Name: misc.StackNameGovernance},
		{Name: misc.StackNameStreaming},
		{Name: misc.StackNameBatch},
		{Name: misc.StackNameProcessing},
		{Name: misc.StackNameConsumption},
		{Name: misc.StackNameObservability},
		{Name: misc.StackNameCosts},
	}

	for _, stack := range stacks {

		fmt.Println(misc.Blue("Deploying %s stack...", stack.Name))

		if err := DeployStack(&stack, templates); err != nil {
			return fmt.Errorf("failed to deploy stack %s: %w", stack.Name, err)
		}

	}

	return nil

}

func DeployStack(stack *Stack, templates embed.FS) error {

	var (
		err      error
		body     []byte
		identity *sts.GetCallerIdentityOutput
	)

	cfClient := cloudformation.NewFromConfig(awsConfig)

	fullStackName := stack.FullStackName()

	if body, err = stack.GetTemplateBody(templates); err != nil {
		return fmt.Errorf("failed to read template content for stack %s: %w", fullStackName, err)
	}

	// Resolve DeployerArn if security stack
	if stack.Name == misc.StackNameSecurity {
		if identity, err = GetCallerIdentity(context.TODO(), GetAWSConfig()); err != nil {
			return fmt.Errorf("failed to get deployer ARN for security stack: %w", err)
		}
		stack.Params = append(stack.Params, types.Parameter{
			ParameterKey:   aws.String("DeployerArn"),
			ParameterValue: identity.Arn,
		})
	}

	createInput := &cloudformation.CreateStackInput{
		StackName:    aws.String(fullStackName),
		TemplateBody: aws.String(string(body)),
		Parameters:   stack.Params,
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
			types.CapabilityCapabilityNamedIam,
		},
	}

	if _, err = cfClient.CreateStack(context.TODO(), createInput); err != nil {
		if strings.Contains(err.Error(), "AlreadyExists") {
			fmt.Println(misc.Yellow("Stack %s already exists. Attempting to update...", fullStackName))
			updateInput := &cloudformation.UpdateStackInput{
				StackName:    aws.String(fullStackName),
				TemplateBody: aws.String(string(body)),
				Parameters:   stack.Params,
				Capabilities: []types.Capability{
					types.CapabilityCapabilityIam,
					types.CapabilityCapabilityNamedIam,
				},
			}
			if _, err = cfClient.UpdateStack(context.TODO(), updateInput); err != nil {
				if strings.Contains(err.Error(), "No updates are to be performed") {
					fmt.Println(misc.Green("No updates needed for stack %s", fullStackName))
					return nil
				}
				return fmt.Errorf("failed to update stack %s: %w", fullStackName, err)
			}
		} else {
			return fmt.Errorf("failed to create stack %s: %w", fullStackName, err)
		}
	}

	fmt.Println(misc.Green("Waiting for stack %s to complete...", fullStackName))
	waiter := cloudformation.NewStackCreateCompleteWaiter(cfClient)
	describeInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(fullStackName),
	}
	if err = waiter.Wait(context.TODO(), describeInput, 30*time.Minute); err != nil {
		return fmt.Errorf("stack %s did not reach complete state: %w", fullStackName, err)
	}

	fmt.Println(misc.Green("Stack %s completed successfully", fullStackName))
	return nil

}

func GetStackOutputs(stack *Stack) (map[string]string, error) {

	cfClient := cloudformation.NewFromConfig(awsConfig)

	resp, err := cfClient.DescribeStacks(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: aws.String(stack.FullStackName()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe stack %s: %w", stack.FullStackName(), err)
	}

	if len(resp.Stacks) == 0 {
		return nil, fmt.Errorf("no stacks found with name %s", stack.FullStackName())
	}

	outputs := make(map[string]string)
	for _, o := range resp.Stacks[0].Outputs {
		if o.OutputKey != nil && o.OutputValue != nil {
			outputs[*o.OutputKey] = *o.OutputValue
		}
	}

	return outputs, nil

}
