package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/mock"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSyncCatalogCreatesBronzeTable(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGlue := mock.NewMockGlueClient(ctrl)
	mockGlue.EXPECT().
		GetTable(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("EntityNotFoundException"))

	mockGlue.EXPECT().
		CreateTable(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, input *glue.CreateTableInput, optFns ...func(*glue.Options)) (*glue.CreateTableOutput, error) {
			assert.Equal(t, "dm_bronze", *input.DatabaseName)
			assert.Equal(t, "brewery", *input.TableInput.Name)
			return &glue.CreateTableOutput{}, nil
		})

	mock.InjectGlueClient(mockGlue)

	mock.InjectStackOutputFetcher(func(cfg aws.Config, key string) (string, error) {
		return "dm-bucket", nil
	})

	db, err := mock.OpenDuckDB()
	require.NoError(t, err)

	require.NoError(t, mock.SeedFakePgCatalogWithSchema(db, "dm_core"))

	err = core.SyncCatalogFromDatabaseSchema(db, "bronze", "brewery")
	assert.NoError(t, err)

}
