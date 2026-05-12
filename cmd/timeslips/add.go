package timeslips

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/spf13/cobra"

	"freeagent/internal/api"
	"freeagent/internal/output"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new timeslip",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, _ := cmd.Flags().GetString("user")
		projectID, _ := cmd.Flags().GetString("project")
		taskID, _ := cmd.Flags().GetString("task")
		hours, _ := cmd.Flags().GetString("hours")
		date, _ := cmd.Flags().GetString("date")
		comment, _ := cmd.Flags().GetString("comment")
		asJSON, _ := cmd.Flags().GetBool("json")

		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		return AddTimeslip(client, userID, projectID, taskID, hours, date, comment, asJSON)
	},
}

func init() {
	addCmd.Flags().String("user", "", "user ID")
	addCmd.Flags().String("project", "", "project ID")
	addCmd.Flags().String("task", "", "task ID")
	addCmd.Flags().String("hours", "", "hours worked, e.g. 2 or 1.5")
	addCmd.Flags().String("date", "", "date (YYYY-MM-DD, defaults to today)")
	addCmd.Flags().String("comment", "", "optional comment")
	addCmd.Flags().Bool("json", false, "output as JSON")

	addCmd.MarkFlagRequired("user")
	addCmd.MarkFlagRequired("project")
	addCmd.MarkFlagRequired("task")
	addCmd.MarkFlagRequired("hours")

	Cmd.AddCommand(addCmd)
}

func AddTimeslip(c *api.Client, userID, projectID, taskID, hours, date, comment string, asJSON bool) error {
	ts := map[string]any{
		"user":     api.BaseURL + "/users/" + userID,
		"project":  api.BaseURL + "/projects/" + projectID,
		"task":     api.BaseURL + "/tasks/" + taskID,
		"dated_on": date,
		"hours":    hours,
	}
	if comment != "" {
		ts["comment"] = comment
	}
	payload := map[string]any{"timeslip": ts}

	var result struct {
		Timeslip struct {
			URL     string `json:"url"`
			DatedOn string `json:"dated_on"`
			Hours   string `json:"hours"`
		} `json:"timeslip"`
	}

	if err := c.APIRequest(http.MethodPost, "/timeslips", payload, &result); err != nil {
		return err
	}

	id := path.Base(result.Timeslip.URL)

	if asJSON {
		return output.JSON(map[string]string{
			"id":       id,
			"dated_on": result.Timeslip.DatedOn,
			"hours":    result.Timeslip.Hours,
		})
	}

	fmt.Printf("Created timeslip %s for %sh on %s\n", id, result.Timeslip.Hours, result.Timeslip.DatedOn)
	return nil
}
