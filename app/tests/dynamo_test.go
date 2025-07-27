package tests

import (
	"context"
	"testing"

	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/require"
)

func TestProcessingControlLifecycle(t *testing.T) {

	MustPersistTestConfig()

	ctx := context.Background()
	cfg := core.GetAWSConfig()

	objectKey := "raw/brewery/testfile.json.gz"

	control, err := core.NewProcessingControl("silver", objectKey)
	require.NoError(t, err)

	control.FileFormat = "json"
	control.FileSize = 12345
	control.RecordCount = 10
	control.ComputeTarget = "test"
	control.Checksum = "abc123"

	err = control.Put(ctx, cfg)
	require.NoError(t, err)

	var retrieved core.ProcessingControl
	err = core.QueryByObjectKey(cfg, ctx, objectKey, &retrieved)
	require.NoError(t, err)
	require.Equal(t, control.ObjectKey, retrieved.ObjectKey)
	require.Equal(t, control.TableName, retrieved.TableName)

	err = control.Delete(ctx, cfg)
	require.NoError(t, err)

}
