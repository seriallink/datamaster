package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/seriallink/datamaster/cli/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Singleton instance and sync primitive for DB connection
var (
	dbConn *gorm.DB
	dbOnce sync.Once
)

// GetConnection returns a singleton *gorm.DB instance, initializing it on first use.
//
// Returns:
//   - *gorm.DB: the initialized database connection.
//   - error: if connection initialization fails.
func GetConnection() (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		err = openConnection()
	})
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// openConnection establishes and persists a connection to an Aurora PostgreSQL instance using GORM.
// Connection details and credentials are retrieved from AWS Secrets Manager and CloudFormation stack outputs.
//
// It uses sync.Once internally to ensure thread-safe one-time initialization.
//
// Returns:
//   - error: if any step in the connection process fails.
func openConnection() error {
	var (
		err             error
		secretArn       string
		secretOutput    *secretsmanager.GetSecretValueOutput
		databaseOutputs map[string]string
	)

	cfg := GetAWSConfig()
	svc := secretsmanager.NewFromConfig(cfg)

	stackSecurity := &Stack{Name: misc.StackNameSecurity}
	stackDatabase := &Stack{Name: misc.StackNameDatabase}

	secretArn, err = stackSecurity.GetStackOutput(cfg, "SecretArn")
	if err != nil {
		return fmt.Errorf("failed to get SecretArn from stack %s: %w", stackSecurity.Name, err)
	}

	secretOutput, err = svc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		return err
	}

	dbSecret := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err = json.Unmarshal([]byte(*secretOutput.SecretString), &dbSecret); err != nil {
		return err
	}

	if databaseOutputs, err = stackDatabase.GetStackOutputs(cfg); err != nil {
		return err
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		url.QueryEscape(dbSecret.Username),
		url.QueryEscape(dbSecret.Password),
		databaseOutputs["WriterEndpoint"],
		databaseOutputs["DatabasePort"],
		databaseOutputs["DatabaseName"])

	dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			LogLevel:      logger.Info,
			Colorful:      true,
			SlowThreshold: 100 * time.Millisecond,
		}),
	})
	if err != nil {
		return err
	}

	// Verifies the connection by pinging the DB.
	instance, _ := dbConn.DB()
	if err = instance.Ping(); err != nil {
		return err
	}

	return nil

}

// CloseConnection gracefully closes the underlying database connection if it's open.
//
// Returns:
//   - error: if fetching the raw connection or closing it fails.
func CloseConnection() error {
	if dbConn != nil {
		sqlDB, err := dbConn.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
