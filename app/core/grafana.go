package core

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	"github.com/aws/aws-sdk-go-v2/service/grafana/types"
)

const serviceAccountName = "dm-service-account"

type GrafanaDatasource struct {
	ID   int64  `json:"id"`
	UID  string `json:"uid"`
	Name string `json:"name"`
}

// PushAllDashboards reads all JSON dashboards from the embedded filesystem and pushes them to Grafana.
//
// Parameters:
//   - fs: the embedded filesystem containing JSON dashboard files under the "dashboards" directory.
//
// Returns:
//   - error: an error if reading the directory or pushing any dashboard fails.
func PushAllDashboards(fs embed.FS) error {

	entries, err := fs.ReadDir("dashboards")
	if err != nil {
		return fmt.Errorf("failed to list dashboards: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		if err = PushDashboard(name, fs); err != nil {
			return fmt.Errorf("failed to push dashboard %s: %w", name, err)
		}
	}

	return nil

}

// PushDashboard uploads a JSON dashboard to Grafana, replacing its datasource placeholder.
//
// Parameters:
//   - name: the name of the dashboard file (without .json extension) located in the "dashboards" folder.
//   - fs: the embedded filesystem containing the dashboard definitions.
//
// Returns:
//   - error: an error if the dashboard file is missing, the datasource is not created, or the upload fails.
func PushDashboard(name string, fs embed.FS) error {

	var (
		err  error
		data []byte
		gds  *GrafanaDatasource
	)

	cfg := GetAWSConfig()

	file := path.Join("dashboards", fmt.Sprintf("%s.json", name))
	data, err = fs.ReadFile(file)
	if err != nil {
		return fmt.Errorf("dashboard file not found: %w", err)
	}

	switch name {
	case "analytics", "costs":
		if gds, err = FindOrCreateAthenaDatasource(cfg); err != nil {
			return fmt.Errorf("failed to ensure Athena datasource: %w", err)
		}
	case "health":
		if gds, err = FindOrCreateCloudWatchDatasource(cfg); err != nil {
			return fmt.Errorf("failed to ensure Athena datasource: %w", err)
		}
	default:
		return fmt.Errorf("unknown dashboard: %s", name)
	}

	data = bytes.ReplaceAll(data, []byte("${Datasource}"), []byte(gds.UID))

	if _, err = doGrafanaRequest(cfg, http.MethodPost, "/api/dashboards/db", data); err != nil {
		return fmt.Errorf("failed to push dashboard %s: %w", name, err)
	}

	fmt.Println(misc.Green("Dashboard '%s' created successfully in Grafana.\n", name))
	return nil

}

// FindOrCreateAthenaDatasource ensures that a Grafana Athena datasource exists and returns its metadata.
//
// If the datasource named "Athena" exists, it is returned. Otherwise, it is created
// using the default configuration with IAM role authentication and the S3 output path
// based on the `ObserverBucketName` stack output.
//
// Parameters:
//   - cfg: AWS configuration used to resolve stack outputs and region.
//
// Returns:
//   - *GrafanaDatasource: the existing or newly created datasource metadata.
//   - error: an error if the operation fails at any step.
func FindOrCreateAthenaDatasource(cfg aws.Config) (*GrafanaDatasource, error) {

	var (
		err    error
		bucket string
		body   []byte
		gds    GrafanaDatasource
	)

	name := "Athena"
	endpoint := "/api/datasources/name/" + name

	bucket, err = (&Stack{Name: misc.StackNameStorage}).GetStackOutput(cfg, "ObserverBucketName")
	if err != nil {
		return nil, err
	}

	body, err = doGrafanaRequest(cfg, http.MethodGet, endpoint, nil)
	if err == nil {
		if err = json.Unmarshal(body, &gds); err != nil {
			return nil, fmt.Errorf("failed to parse existing datasource: %w", err)
		}
		return &gds, nil
	} else if !strings.Contains(err.Error(), "404") {
		return nil, fmt.Errorf("failed to check existing datasource: %w", err)
	}

	payload := map[string]interface{}{
		"name":      name,
		"type":      "grafana-athena-datasource",
		"access":    "proxy",
		"isDefault": true,
		"jsonData": map[string]interface{}{
			"defaultRegion":  cfg.Region,
			"catalog":        "AwsDataCatalog",
			"database":       "dm_gold",
			"workgroup":      "primary",
			"outputLocation": fmt.Sprintf("s3://%s/athena/results/", bucket),
			"authType":       "ec2_iam_role",
		},
	}

	body, err = doGrafanaRequest(cfg, http.MethodPost, "/api/datasources", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create datasource: %w", err)
	}

	if err = json.Unmarshal(body, &gds); err != nil {
		return nil, fmt.Errorf("failed to parse created datasource: %w", err)
	}

	fmt.Println(misc.Green("Athena datasource created successfully."))
	return &gds, nil

}

// FindOrCreateCloudWatchDatasource ensures that a Grafana CloudWatch datasource exists and returns its metadata.
//
// If a datasource named "CloudWatch" already exists in Grafana, it is returned.
// Otherwise, the function creates a new CloudWatch datasource using the default authentication
// method (IAM role from environment) and region specified in the AWS config.
//
// Parameters:
//   - cfg: AWS configuration used to determine the region and credentials.
//
// Returns:
//   - *GrafanaDatasource: the existing or newly created CloudWatch datasource.
//   - error: an error if the operation fails at any step.
func FindOrCreateCloudWatchDatasource(cfg aws.Config) (*GrafanaDatasource, error) {

	var (
		err  error
		body []byte
		gds  GrafanaDatasource
	)

	name := "CloudWatch"
	endpoint := "/api/datasources/name/" + name

	body, err = doGrafanaRequest(cfg, http.MethodGet, endpoint, nil)
	if err == nil {
		if err = json.Unmarshal(body, &gds); err != nil {
			return nil, fmt.Errorf("failed to parse existing datasource: %w", err)
		}
		return &gds, nil
	} else if !strings.Contains(err.Error(), "404") {
		return nil, fmt.Errorf("failed to check existing datasource: %w", err)
	}

	payload := map[string]interface{}{
		"name":      name,
		"type":      "cloudwatch",
		"access":    "proxy",
		"uid":       "cloudwatch",
		"isDefault": false,
		"jsonData": map[string]interface{}{
			"authType":      "default", // usa a role padrão do ambiente (federado ou EC2)
			"defaultRegion": cfg.Region,
		},
	}

	body, err = doGrafanaRequest(cfg, http.MethodPost, "/api/datasources", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create CloudWatch datasource: %w", err)
	}

	if err = json.Unmarshal(body, &gds); err != nil {
		return nil, fmt.Errorf("failed to parse created datasource: %w", err)
	}

	fmt.Println(misc.Green("CloudWatch datasource created successfully."))
	return &gds, nil

}

// CreateGrafanaToken generates a temporary API token for a service account in an AWS-managed Grafana workspace.
//
// It retrieves the workspace ID from the CloudFormation stack, ensures that the required service account exists,
// and then creates a new token with a unique name and the specified TTL (time-to-live).
//
// Parameters:
//   - cfg: AWS configuration used to authenticate and interact with the Grafana service.
//   - ttlSeconds: time-to-live for the token, in seconds.
//
// Returns:
//   - string: the token key that can be used to authenticate Grafana API requests.
//   - error: an error if the token creation fails at any step.
func CreateGrafanaToken(cfg aws.Config, ttlSeconds int32) (string, error) {

	client := grafana.NewFromConfig(cfg)

	workspaceID, err := (&Stack{Name: misc.StackNameObservability}).GetStackOutput(cfg, "GrafanaWorkspaceId")
	if err != nil {
		return "", err
	}

	accountId, err := FindOrCreateServiceAccount(cfg, workspaceID)
	if err != nil {
		return "", err
	}

	tokenName := fmt.Sprintf("%s-token-%d", serviceAccountName, time.Now().Unix())
	tokenOutput, err := client.CreateWorkspaceServiceAccountToken(context.TODO(), &grafana.CreateWorkspaceServiceAccountTokenInput{
		WorkspaceId:      aws.String(workspaceID),
		ServiceAccountId: aws.String(accountId),
		Name:             aws.String(tokenName),
		SecondsToLive:    aws.Int32(ttlSeconds),
	})
	if err != nil {
		return "", fmt.Errorf("create token failed: %w", err)
	}

	return *tokenOutput.ServiceAccountToken.Key, nil

}

// FindOrCreateServiceAccount locates an existing service account in a Grafana workspace or creates a new one if it doesn't exist.
//
// It searches for a service account with the predefined name and returns its ID. If not found, it creates
// a new service account with `admin` privileges.
//
// Parameters:
//   - cfg: AWS configuration used for authenticating Grafana service calls.
//   - workspaceID: ID of the Grafana workspace where the service account should be found or created.
//
// Returns:
//   - string: the ID of the existing or newly created service account.
//   - error: an error if the listing or creation of the service account fails.
func FindOrCreateServiceAccount(cfg aws.Config, workspaceID string) (string, error) {

	client := grafana.NewFromConfig(cfg)

	accounts, err := client.ListWorkspaceServiceAccounts(context.TODO(), &grafana.ListWorkspaceServiceAccountsInput{
		WorkspaceId: aws.String(workspaceID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list service accounts: %w", err)
	}

	for _, sa := range accounts.ServiceAccounts {
		if *sa.Name == serviceAccountName {
			return *sa.Id, nil
		}
	}

	input, err := client.CreateWorkspaceServiceAccount(context.TODO(), &grafana.CreateWorkspaceServiceAccountInput{
		WorkspaceId: aws.String(workspaceID),
		Name:        aws.String(serviceAccountName),
		GrafanaRole: types.RoleAdmin,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create service account: %w", err)
	}

	return *input.Id, nil

}

// doGrafanaRequest performs an authenticated HTTP request to the Grafana workspace API.
//
// It retrieves the workspace endpoint and creates a temporary token with 60 seconds of TTL.
// The request is made with the provided HTTP method, path, and payload. The payload can be
// a raw JSON byte slice or a Go value that will be marshaled into JSON.
//
// Parameters:
//   - cfg: AWS configuration used to access stack outputs and Grafana token.
//   - method: HTTP method (e.g., "GET", "POST").
//   - path: Relative API path (e.g., "/api/datasources").
//   - payload: Request body. Can be a raw `[]byte` or a struct/map to be marshaled into JSON.
//
// Returns:
//   - []byte: The response body from Grafana.
//   - error: An error if any step (token, request, response) fails.
func doGrafanaRequest(cfg aws.Config, method, path string, payload interface{}) ([]byte, error) {

	var (
		err      error
		endpoint string
		token    string
		body     io.Reader
		buf      []byte
		request  *http.Request
		response *http.Response
		content  []byte
	)

	endpoint, err = (&Stack{Name: misc.StackNameObservability}).GetStackOutput(cfg, "GrafanaWorkspaceEndpoint")
	if err != nil {
		return nil, fmt.Errorf("failed to get Grafana endpoint: %w", err)
	}
	url := "https://" + strings.TrimSuffix(endpoint, "/") + path

	token, err = CreateGrafanaToken(cfg, 60)
	if err != nil {
		return nil, fmt.Errorf("failed to get Grafana token: %w", err)
	}

	if _, ok := payload.([]byte); ok {
		body = bytes.NewReader(payload.([]byte))
	} else {
		if payload != nil {
			buf, err = json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal payload: %w", err)
			}
			body = bytes.NewReader(buf)
		}
	}

	request, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	response, err = http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	content, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Grafana response: %w", err)
	}

	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("api error (%d): %s", response.StatusCode, string(content))
	}

	return content, nil

}
