package tests

import (
	"bytes"
	"testing"

	"github.com/seriallink/datamaster/app/bronze"
	"github.com/seriallink/datamaster/app/core"

	"github.com/stretchr/testify/require"
)

func TestWriteParquet(t *testing.T) {

	records := []bronze.Brewery{
		{BreweryId: 1, BreweryName: "Angry Chair", Operation: "insert"},
		{BreweryId: 2, BreweryName: "Side Project", Operation: "insert"},
	}

	var buf bytes.Buffer
	err := core.WriteParquet(records, &buf)
	require.NoError(t, err)
	require.True(t, buf.Len() > 0)

}
