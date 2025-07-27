package tests

import (
	"strings"
	"testing"

	"github.com/seriallink/datamaster/app/bronze"
	"github.com/seriallink/datamaster/app/core"
	"github.com/seriallink/datamaster/app/model"

	"github.com/stretchr/testify/require"
)

func TestValidateRow(t *testing.T) {
	row := map[string]any{
		"brewery_id":   1,
		"brewery_name": "3 Sons",
	}
	err := core.ValidateRow(row, &model.Brewery{}, 1)
	require.NoError(t, err)
}

func TestValidateRowMissingName(t *testing.T) {
	row := map[string]any{
		"brewery_id": 1,
		// brewery_name missing
	}
	err := core.ValidateRow(row, &bronze.Brewery{}, 3)
	require.ErrorContains(t, err, "missing field")
}

func TestCsvToMap(t *testing.T) {

	input := `brewery_id,brewery_name,operation
1,Mikkeller,insert
2,Tree House,insert
`
	reader := strings.NewReader(input)
	rows, err := core.CsvToMap(reader, &bronze.Brewery{})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, "Mikkeller", rows[0]["brewery_name"])
	require.Equal(t, "Tree House", rows[1]["brewery_name"])

}

func TestUnmarshalRecords(t *testing.T) {

	data := []map[string]any{
		{"brewery_id": 1, "brewery_name": "Burial"},
		{"brewery_id": 2, "brewery_name": "Adroit Theory"},
	}

	out, err := core.UnmarshalRecords(&bronze.Brewery{}, data)
	require.NoError(t, err)

	list, ok := out.(*[]bronze.Brewery)
	require.True(t, ok)
	require.Len(t, *list, 2)
	require.Equal(t, int64(1), (*list)[0].BreweryId)
	require.Equal(t, "Adroit Theory", (*list)[1].BreweryName)

}
