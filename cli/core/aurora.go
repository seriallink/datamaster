package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AuroraSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func ConnectAuroraFromSecret(secretArn string) (*sql.DB, error) {

	var (
		err             error
		auroraSecret    AuroraSecret
		secretOutput    *secretsmanager.GetSecretValueOutput
		databaseOutputs map[string]string
	)

	svc := secretsmanager.NewFromConfig(GetAWSConfig())

	secretOutput, err = svc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{SecretId: aws.String(secretArn)})
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(*secretOutput.SecretString), &auroraSecret); err != nil {
		return nil, err
	}

	databaseOutputs, err = GetStackOutputs(&Stack{Name: misc.StackNameDatabase})
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		url.QueryEscape(auroraSecret.Username),
		url.QueryEscape(auroraSecret.Password),
		databaseOutputs["WriterEndpoint"],
		databaseOutputs["DatabasePort"],
		databaseOutputs["DatabaseName"])

	return sql.Open("postgres", dsn)

}
