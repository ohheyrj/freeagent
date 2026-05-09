package main

import (
	"encoding/json"
	"github.com/olekukonko/tablewriter"
	"os"
)

func printTable(headers []string, rows [][]string) error {
	t := tablewriter.NewWriter(os.Stdout)
	t.Header(headers)
	if err := t.Bulk(rows); err != nil {
		return err
	}
	return t.Render()
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
