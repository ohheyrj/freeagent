package invoices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/ohheyrj/freeagent/internal/api"
	"github.com/spf13/cobra"
)

var timeslipGrouping = map[string]string{
	"single":   "billed_grouped_by_single_timeslip",
	"separate": "billed_grouped_by_timeslip",
	"task":     "billed_grouped_by_timeslip_task",
	"date":     "billed_grouped_by_timeslip_date",
}

type createOptions struct {
	project, date, currency, reference, comments, grouping string
	terms                                                  int
	asJSON, dryRun                                         bool
}

type invoiceInput struct {
	Contact              string `json:"contact"`
	Project              string `json:"project"`
	DatedOn              string `json:"dated_on"`
	PaymentTermsInDays   int    `json:"payment_terms_in_days"`
	Currency             string `json:"currency"`
	Reference            string `json:"reference,omitempty"`
	Comments             string `json:"comments,omitempty"`
	IncludeTimeslips     string `json:"include_timeslips"`
	IncludeExpenses      any    `json:"include_expenses"`
	IncludeEstimates     any    `json:"include_estimates"`
	SendNewInvoiceEmails bool   `json:"send_new_invoice_emails"`
}

func newCreateCommand(newClient func() (*api.Client, error)) *cobra.Command {
	var opts createOptions
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"add"},
		Short:   "Create a draft invoice from a project's unbilled timeslips",
		Long: "Create a draft invoice using the project's contact and currency.\n" +
			"Includes unbilled timeslips on one line by default; excludes expenses and estimates.\n" +
			"The invoice date is not a timeslip date filter. This command does not send the invoice.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !regexp.MustCompile(`^[0-9]+$`).MatchString(opts.project) || strings.TrimLeft(opts.project, "0") == "" {
				return fmt.Errorf("--project must be a positive numeric project ID")
			}
			date := opts.date
			if date == "" {
				date = time.Now().Format(time.DateOnly)
			}
			if _, err := time.Parse(time.DateOnly, date); err != nil {
				return fmt.Errorf("--date must be a valid date in YYYY-MM-DD format")
			}
			if opts.terms < 0 {
				return fmt.Errorf("--payment-terms must be zero or greater")
			}
			grouping, ok := timeslipGrouping[opts.grouping]
			if !ok {
				return fmt.Errorf("--include-timeslips must be single, separate, task, or date")
			}
			currency := strings.ToUpper(opts.currency)
			if currency != "" && !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(currency) {
				return fmt.Errorf("--currency must be a three-letter currency code")
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			projectURL := strings.TrimRight(api.BaseURL, "/") + "/projects/" + opts.project
			var project struct {
				Project struct {
					Contact  string `json:"contact"`
					Currency string `json:"currency"`
				} `json:"project"`
			}
			if err := client.APIRequest(http.MethodGet, "/projects/"+opts.project, nil, &project); err != nil {
				return fmt.Errorf("fetch project: %w", err)
			}
			if project.Project.Contact == "" {
				return fmt.Errorf("project %s has no contact", opts.project)
			}
			if currency == "" {
				currency = project.Project.Currency
			}
			if currency == "" {
				return fmt.Errorf("project %s has no currency; specify --currency", opts.project)
			}
			payload := struct {
				Invoice invoiceInput `json:"invoice"`
			}{Invoice: invoiceInput{
				Contact: project.Project.Contact, Project: projectURL,
				DatedOn: date, PaymentTermsInDays: opts.terms, Currency: currency,
				Reference: opts.reference, Comments: opts.comments, IncludeTimeslips: grouping,
			}}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			if opts.dryRun {
				return enc.Encode(payload)
			}
			var result struct {
				Invoice map[string]any `json:"invoice"`
			}
			if err := client.APIRequest(http.MethodPost, "/invoices", payload, &result); err != nil {
				return fmt.Errorf("create invoice: %w", err)
			}
			invoiceURL, _ := result.Invoice["url"].(string)
			if invoiceURL == "" {
				return fmt.Errorf("invoice creation returned no invoice URL; check FreeAgent before retrying")
			}
			id := path.Base(invoiceURL)
			if opts.asJSON {
				result.Invoice["id"] = id
				return enc.Encode(result.Invoice)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created invoice %s (reference %v, status %v): %v %v\n",
				id, result.Invoice["reference"], result.Invoice["status"], result.Invoice["total_value"], result.Invoice["currency"])
			return err
		},
	}
	cmd.Flags().StringVar(&opts.project, "project", "", "project ID (uses its contact and currency)")
	cmd.Flags().StringVar(&opts.date, "date", "", "invoice date (YYYY-MM-DD, defaults to today; does not filter timeslips)")
	cmd.Flags().IntVar(&opts.terms, "payment-terms", 30, "payment terms in days (0 means due on receipt)")
	cmd.Flags().StringVar(&opts.currency, "currency", "", "currency code (defaults to project currency)")
	cmd.Flags().StringVar(&opts.reference, "reference", "", "invoice reference (defaults to FreeAgent's next reference)")
	cmd.Flags().StringVar(&opts.comments, "comments", "", "additional text at the bottom of the invoice")
	cmd.Flags().StringVar(&opts.grouping, "include-timeslips", "single", "group unbilled timeslips: single|separate|task|date")
	cmd.Flags().BoolVar(&opts.asJSON, "json", false, "output the created invoice as JSON")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "fetch project and print request JSON without creating an invoice")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}
