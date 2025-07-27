package tests

import (
	"context"
	"testing"

	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/require"
)

func TestGovernanceMaskPIIData(t *testing.T) {

	MustPersistTestConfig()

	cfg := core.GetAWSConfig()
	ctx := context.TODO()

	data := []map[string]any{
		{"name": "John Doe", "email": "john_doe@example.com"},
		{"name": "Alice Johnson", "email": "alice@example.com"},
	}

	masked, err := core.MaskPIIData(cfg, ctx, data)
	require.NoError(t, err)
	t.Logf("Masked result: %+v", masked)

}

func TestGovernanceMaskValue(t *testing.T) {

	tests := []struct {
		input  any
		expect string
	}{
		{"", ""},
		{"A", "*"},
		{"AB", "**"},
		{"1234", "****"},
		{"secret", "s****t"},
		{"1234567890", "1********0"},
	}

	for _, tt := range tests {
		masked := core.MaskValue(tt.input)
		require.Equal(t, tt.expect, masked)
	}

}

func TestGovernanceSampleData(t *testing.T) {

	data := []map[string]any{
		{"name": "Alice"},
		{"name": "Bob"},
		{"name": "Carol"},
	}

	sampled := core.SampleData(data, 2)
	require.Len(t, sampled, 2)
	require.Equal(t, "Alice", sampled[0]["name"])
	require.Equal(t, "Bob", sampled[1]["name"])

	all := core.SampleData(data, 5)
	require.Len(t, all, 3)

}
