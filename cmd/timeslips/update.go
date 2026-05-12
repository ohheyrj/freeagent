package timeslips

import (
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/spf13/cobra"

	"freeagent/internal/api"
	"freeagent/internal/output"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing timeslip",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if _, err := strconv.Atoi(id); err != nil {
			return fmt.Errorf("invalid timeslip ID %q: must be numeric", id)
		}

		asJSON, _ := cmd.Flags().GetBool("json")
		client, err := api.NewClient()
		if err != nil {
			return err
		}
		return updateTimeslip(client, id, cmd, asJSON)
	},
}

func init() {
	updateCmd.Flags().String("user", "", "change user ID")
	updateCmd.Flags().String("project", "", "change project ID")
	updateCmd.Flags().String("task", "", "change task ID")
	updateCmd.Flags().String("hours", "", "change hours worked, e.g. 2 or 1.5")
	updateCmd.Flags().String("date", "", "change date (YYYY-MM-DD)")
	updateCmd.Flags().String("comment", "", "change comment")
	updateCmd.Flags().Bool("json", false, "output as JSON")

	updateCmd.MarkFlagsOneRequired(
		"user", "project", "task", "hours", "date", "comment",
	)

	Cmd.AddCommand(updateCmd)
}

func updateTimeslip(c *api.Client, id string, cmd *cobra.Command, asJSON bool) error {
	ts := map[string]any{}

	if cmd.Flags().Changed("user") {
		v, _ := cmd.Flags().GetString("user")
		ts["user"] = api.BaseURL + "/users/" + v
	}
	if cmd.Flags().Changed("project") {
		v, _ := cmd.Flags().GetString("project")
		ts["project"] = api.BaseURL + "/projects/" + v
	}
	if cmd.Flags().Changed("task") {
		v, _ := cmd.Flags().GetString("task")
		ts["task"] = api.BaseURL + "/tasks/" + v
	}
	if cmd.Flags().Changed("hours") {
		v, _ := cmd.Flags().GetString("hours")
		ts["hours"] = v
	}
	if cmd.Flags().Changed("date") {
		v, _ := cmd.Flags().GetString("date")
		ts["dated_on"] = v
	}
	if cmd.Flags().Changed("comment") {
		v, _ := cmd.Flags().GetString("comment")
		ts["comment"] = v
	}
	payload := map[string]any{"timeslip": ts}

	var result struct {
		Timeslip struct {
			URL     string `json:"url"`
			DatedOn string `json:"dated_on"`
			Hours   string `json:"hours"`
		} `json:"timeslip"`
	}
	endpoint := "/timeslips/" + id
	if err := c.APIRequest(http.MethodPut, endpoint, payload, &result); err != nil {
		return err
	}
	rid := path.Base(result.Timeslip.URL)

	if asJSON {
		return output.JSON(map[string]string{
			"id":       rid,
			"dated_on": result.Timeslip.DatedOn,
			"hours":    result.Timeslip.Hours,
		})
	}

	fmt.Printf("Updated timeslip %s (%sh on %s)\n", rid, result.Timeslip.Hours, result.Timeslip.DatedOn)
	return nil
}
