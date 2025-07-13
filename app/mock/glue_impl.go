package mock

import (
	"context"

	"github.com/seriallink/datamaster/app/core"

	"github.com/aws/aws-sdk-go-v2/service/glue"
)

var glueClient GlueClient = glue.NewFromConfig(core.GetAWSConfig())

type GlueClient interface {
	GetTable(ctx context.Context, input *glue.GetTableInput, optFns ...func(*glue.Options)) (*glue.GetTableOutput, error)
	CreateTable(ctx context.Context, input *glue.CreateTableInput, optFns ...func(*glue.Options)) (*glue.CreateTableOutput, error)
	UpdateTable(ctx context.Context, input *glue.UpdateTableInput, optFns ...func(*glue.Options)) (*glue.UpdateTableOutput, error)
}

func InjectGlueClient(mock GlueClient) {
	glueClient = mock
}
