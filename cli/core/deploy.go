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

// Stack represents a CloudFormation stack with optional prefix and parameters.
type Stack struct {
	Name   string            // Logical name of the stack (e.g., "network", "roles").
	Prefix string            // Optional prefix to be prepended to the stack name.
	Params []types.Parameter // Parameters to pass to the CloudFormation template.
}

// FullStackName returns the full stack name by concatenating the prefix (or default) and name.
func (s *Stack) FullStackName() string {
	prefix := s.Prefix
	if prefix == "" {
		prefix = misc.DefaultProjectPrefix
	}
	return fmt.Sprintf("%s-%s", prefix, s.Name)
}

// TemplateFilePath returns the relative path of the CloudFormation template file for the stack.
func (s *Stack) TemplateFilePath() string {
	return misc.TemplatesPath + "/" + s.Name + misc.TemplateExtension
}

// GetTemplateBody reads and returns the contents of the embedded CloudFormation template for the stack.
//
// Parameters:
//   - templates: the embedded file system containing all templates.
//
// Returns:
//   - []byte: the contents of the template file.
//   - error: if reading the file fails.
func (s *Stack) GetTemplateBody(templates embed.FS) ([]byte, error) {
	templatePath := s.TemplateFilePath()
	templateBody, err := templates.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}
	return templateBody, nil
}

// DeployAllStacks deploys all the predefined CloudFormation stacks in logical order.
// It uses DeployStack internally and stops at the first encountered failure.
//
// Parameters:
//   - templates: the embedded file system with all template definitions.
//
// Returns:
//   - error: if any stack deployment fails.
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

// DeployStack deploys or updates a specific CloudFormation stack using the provided template.
//
// It handles creation, update (when already exists), and waits for completion.
//
// Parameters:
//   - stack: the stack definition (name, params, etc.).
//   - templates: embedded file system containing template files.
//
// Returns:
//   - error: if the stack creation or update fails.
func DeployStack(stack *Stack, templates embed.FS) error {

	var (
		err      error
		body     []byte
		identity *sts.GetCallerIdentityOutput
	)

	cfClient := cloudformation.NewFromConfig(GetAWSConfig())
	fullStackName := stack.FullStackName()

	if body, err = stack.GetTemplateBody(templates); err != nil {
		return fmt.Errorf("failed to read template content for stack %s: %w", fullStackName, err)
	}

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
				Capabilities: createInput.Capabilities,
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

// GetStackOutputs retrieves all the outputs from a given CloudFormation stack.
//
// Parameters:
//   - stack: the stack to query.
//
// Returns:
//   - map[string]string: output key-value pairs from the stack.
//   - error: if the stack cannot be found or described.
func GetStackOutputs(stack *Stack) (map[string]string, error) {

	cfClient := cloudformation.NewFromConfig(GetAWSConfig())

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

// GetStackOutput retrieves a specific output value from a given CloudFormation stack.
//
// Parameters:
//   - stack: the stack to query.
//   - key: the name of the output key to retrieve.
//
// Returns:
//   - string: the value associated with the given output key.
//   - error: if the stack cannot be described or the output key is not found.
func GetStackOutput(stack *Stack, key string) (string, error) {

	outputs, err := GetStackOutputs(stack)
	if err != nil {
		return "", fmt.Errorf("failed to get outputs for stack %s: %w", stack.Name, err)
	}

	value := outputs[key]
	if value == "" {
		return "", fmt.Errorf("output %q not found in stack %s", key, stack.Name)
	}

	return value, nil

}

// ExtractNameFromArn extracts the resource name from a full AWS ARN.
func ExtractNameFromArn(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return arn
}
