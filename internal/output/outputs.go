package output

import (
	"encoding/json"
	"os"

	"github.com/olekukonko/tablewriter"
)

func Table(headers []string, rows [][]string) error {
	t := tablewriter.NewWriter(os.Stdout)
	t.Header(headers)
	if err := t.Bulk(rows); err != nil {
		return err
	}
	return t.Render()
}

func JSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
