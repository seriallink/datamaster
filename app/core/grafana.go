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
	case "analytics":
		if gds, err = FindOrCreateAthenaDatasource(cfg); err != nil {
			return fmt.Errorf("failed to ensure Athena datasource: %w", err)
		}
	case "health", "costs":
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
