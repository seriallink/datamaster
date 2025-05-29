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
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Stack represents a CloudFormation stack.
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

// GetStackOutputs retrieves all output key-value pairs from the specified CloudFormation stack.
//
// It describes the stack using its full name and returns a map of outputs,
// where each key is the output name and each value is the output value.
//
// Returns:
//   - map[string]string: output values from the stack.
//   - error: if the stack cannot be found or described.
func (s *Stack) GetStackOutputs() (map[string]string, error) {

	cfClient := cloudformation.NewFromConfig(GetAWSConfig())

	resp, err := cfClient.DescribeStacks(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.FullStackName()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe stack %s: %w", s.FullStackName(), err)
	}

	if len(resp.Stacks) == 0 {
		return nil, fmt.Errorf("no stacks found with name %s", s.FullStackName())
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
func (s *Stack) GetStackOutput(key string) (string, error) {

	outputs, err := s.GetStackOutputs()
	if err != nil {
		return "", fmt.Errorf("failed to get outputs for stack %s: %w", s.Name, err)
	}

	value := outputs[key]
	if value == "" {
		return "", fmt.Errorf("output %q not found in stack %s", key, s.Name)
	}

	return value, nil

}

// DeployAllStacks deploys all the predefined CloudFormation stacks in logical order.
// It uses DeployStack internally and stops at the first encountered failure.
//
// Parameters:
//   - templates: the embedded file system with all template definitions.
//
// Returns:
//   - error: if any stack deployment fails.
func DeployAllStacks(templates, artifacts embed.FS) error {

	stacks := []Stack{
		{Name: misc.StackNameNetwork},
		{Name: misc.StackNameRoles},
		{Name: misc.StackNameSecurity},
		{Name: misc.StackNameStorage},
		{Name: misc.StackNameFunctions},
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
		if err := DeployStack(&stack, templates, artifacts); err != nil {
			return fmt.Errorf("failed to deploy stack %s: %w", stack.Name, err)
		}
	}

	return nil

}

// DeployStack deploys or updates a specific CloudFormation stack using the provided template.
//
// It handles creation, update (when it already exists), and waits for completion.
//
// Parameters:
//   - stack: the stack definition (name, params, etc.).
//   - templates: embedded file system containing template files.
//
// Returns:
//   - error: if the stack creation or update fails.
func DeployStack(stack *Stack, templates, artifacts embed.FS) error {

	cfg := GetAWSConfig()
	cfClient := cloudformation.NewFromConfig(cfg)
	fullStackName := stack.FullStackName()

	templateBody, err := stack.GetTemplateBody(templates)
	if err != nil {
		return fmt.Errorf("failed to read template for stack %s: %w", fullStackName, err)
	}

	if err = injectParameters(stack, cfg); err != nil {
		return err
	}

	if err = prepareArtifacts(stack, artifacts); err != nil {
		return err
	}

	if err = createOrUpdateStack(stack, cfClient, string(templateBody)); err != nil {
		return err
	}

	if err = waitForCompletion(fullStackName, cfClient); err != nil {
		return err
	}

	if err = postDeploymentHooks(stack, cfg); err != nil {
		return err
	}

	fmt.Println(misc.Green("Stack %s completed successfully", fullStackName))
	return nil

}

// injectParameters injects additional parameters into a CloudFormation stack definition, based on stack-specific rules.
//
// Currently, it appends the caller's ARN as the "DeployerArn" parameter if the stack name is "security".
// For all other stacks, no parameters are added.
//
// Parameters:
//   - stack: pointer to the stack being prepared for deployment.
//   - cfg: initialized AWS configuration used to retrieve the caller's identity.
//
// Returns:
//   - error: if the deployer's identity cannot be retrieved.
func injectParameters(stack *Stack, cfg aws.Config) error {
	if stack.Name != misc.StackNameSecurity {
		return nil
	}
	identity, err := GetCallerIdentity(context.TODO(), cfg)
	if err != nil {
		return fmt.Errorf("failed to get deployer ARN for security stack: %w", err)
	}
	stack.Params = append(stack.Params, types.Parameter{
		ParameterKey:   aws.String("DeployerArn"),
		ParameterValue: identity.Arn,
	})
	return nil
}

// prepareArtifacts uploads embedded artifacts required by a specific CloudFormation stack.
//
// Currently, it uploads Lambda deployment packages if the stack name is "functions".
// For all other stacks, this function is a no-op.
//
// Parameters:
//   - stack: pointer to the stack being deployed.
//   - artifacts: embedded file system containing all artifacts (e.g., Lambda zips).
//
// Returns:
//   - error: if uploading the artifacts fails.
func prepareArtifacts(stack *Stack, artifacts embed.FS) error {
	if stack.Name != misc.StackNameFunctions {
		return nil
	}
	fmt.Println(misc.Blue("Uploading Lambda artifacts..."))
	if err := UploadArtifacts(artifacts); err != nil {
		return fmt.Errorf("failed to upload lambda artifacts: %w", err)
	}
	return nil
}

// createOrUpdateStack attempts to create a CloudFormation stack,
// or updates it if it already exists.
//
// If the stack already exists, it calls UpdateStack with the same template body and parameters.
// If no changes are detected during the update, it exits gracefully.
//
// Parameters:
//   - stack: pointer to the stack definition, including parameters.
//   - cfClient: initialized CloudFormation client.
//   - body: rendered CloudFormation template body as a string.
//
// Returns:
//   - error: if stack creation or update fails, or if an unexpected condition occurs.
func createOrUpdateStack(stack *Stack, cfClient *cloudformation.Client, body string) error {

	fullName := stack.FullStackName()

	createInput := &cloudformation.CreateStackInput{
		StackName:    aws.String(fullName),
		TemplateBody: aws.String(body),
		Parameters:   stack.Params,
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
			types.CapabilityCapabilityNamedIam,
		},
	}

	_, err := cfClient.CreateStack(context.TODO(), createInput)
	if err == nil {
		return nil
	}

	if !strings.Contains(err.Error(), "AlreadyExists") {
		return fmt.Errorf("failed to create stack %s: %w", fullName, err)
	}

	fmt.Println(misc.Yellow("Stack %s already exists. Attempting to update...", fullName))
	updateInput := &cloudformation.UpdateStackInput{
		StackName:    aws.String(fullName),
		TemplateBody: aws.String(body),
		Parameters:   stack.Params,
		Capabilities: createInput.Capabilities,
	}
	_, err = cfClient.UpdateStack(context.TODO(), updateInput)
	if err != nil {
		if strings.Contains(err.Error(), "No updates are to be performed") {
			fmt.Println(misc.Green("No updates needed for stack %s", fullName))
			return nil
		}
		return fmt.Errorf("failed to update stack %s: %w", fullName, err)
	}

	return nil

}

// waitForCompletion waits for a CloudFormation stack to reach the CREATE_COMPLETE state.
//
// It uses the StackCreateComplete waiter and times out after 30 minutes.
//
// Parameters:
//   - stackName: name of the stack to monitor.
//   - cfClient: initialized CloudFormation client.
//
// Returns:
//   - error: if the stack fails to complete successfully or the wait times out.
func waitForCompletion(stackName string, cfClient *cloudformation.Client) error {
	fmt.Println(misc.Green("Waiting for stack %s to complete...", stackName))
	waiter := cloudformation.NewStackCreateCompleteWaiter(cfClient)
	describeInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	err := waiter.Wait(context.TODO(), describeInput, 30*time.Minute)
	if err != nil {
		return fmt.Errorf("stack %s did not reach complete state: %w", stackName, err)
	}
	return nil
}

// postDeploymentHooks executes custom logic after a stack has been successfully deployed.
//
// Currently, it configures the S3-to-Lambda notification if the deployed stack is "functions".
// For all other stacks, it performs no actions.
//
// Parameters:
//   - stack: pointer to the deployed stack.
//   - cfg: AWS configuration used to initialize required service clients.
//
// Returns:
//   - error: if any post-deployment action fails.
func postDeploymentHooks(stack *Stack, cfg aws.Config) error {
	if stack.Name == misc.StackNameFunctions {
		fmt.Println(misc.Blue("Configuring S3 → Lambda notification..."))
		err := ConfigureS3Notification(context.TODO(), s3.NewFromConfig(cfg), lambda.NewFromConfig(cfg))
		if err != nil {
			return fmt.Errorf("failed to configure S3 notification: %w", err)
		}
	}
	return nil
}

// ExtractNameFromArn extracts the resource name from a full AWS ARN.
func ExtractNameFromArn(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return arn
}
