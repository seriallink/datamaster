package tests

import (
	"embed"
	"testing"

	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/require"
)

//go:embed stacks/test.yml
var templates embed.FS

func TestDeployStack(t *testing.T) {
	MustPersistTestConfig()
	stack := &core.Stack{
		Name: "test",
	}
	err := core.DeployStack(stack, templates, templates, templates)
	require.NoError(t, err)
}
