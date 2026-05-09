package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	command, args := args[0], args[1:]

	if command == "help" || command == "-h" || command == "--help" {
		printUsage()
		return nil
	}

	client, err := newFaClient()
	if err != nil {
		return err
	}

	switch command {
	case "projects":
		fs := flag.NewFlagSet("projects", flag.ContinueOnError)
		view := fs.String("view", "", "filter: active|completed|cancelled|hidden")
		asJSON := fs.Bool("json", false, "output as JSON")
		if err := fs.Parse(args); err != nil {
			return err
		}
		return cmdListProjects(client, *view, *asJSON)
	case "tasks":
		fs := flag.NewFlagSet("tasks", flag.ContinueOnError)
		projectID := fs.String("projectID", "", "valid project ID")
		asJSON := fs.Bool("json", false, "output as JSON")
		if err := fs.Parse(args); err != nil {
			return err
		}
		return cmdListTasks(client, *projectID, *asJSON)
	case "users":
		fs := flag.NewFlagSet("users", flag.ContinueOnError)
		asJSON := fs.Bool("json", false, "output as JSON")
		if err := fs.Parse(args); err != nil {
			return err
		}
		return cmdListUsers(client, *asJSON)
	case "timeslips":
		fs := flag.NewFlagSet("timeslips", flag.ContinueOnError)
		var opts listTimeslipOpts
		fs.StringVar(&opts.View, "view", "", "filter: all|recent|unbilled|running")
		fs.StringVar(&opts.UserID, "user", "", "filter by user ID")
		fs.StringVar(&opts.ProjectID, "project", "", "filter by project ID")
		fs.StringVar(&opts.TaskID, "task", "", "filter by task ID")
		fs.StringVar(&opts.FromDate, "from", "", "from date (YYYY-MM-DD)")
		fs.StringVar(&opts.ToDate, "to", "", "to date (YYYY-MM-DD)")
		asJSON := fs.Bool("json", false, "output as JSON")
		if err := fs.Parse(args); err != nil {
			return err
		}
		return cmdListTimeslips(client, opts, *asJSON)
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: freeagent <command> [flags]

Commands:
  projects    List projects        (--view, --json)
  tasks       List tasks           (--projectID, --json)
  users       List users           (--json)
  timeslips   List timeslips       (--view, --user, --project, --task, --from, --to, --json)
  help        Show this help

Run 'freeagent <command> --help' for command-specific flags.`)
}
