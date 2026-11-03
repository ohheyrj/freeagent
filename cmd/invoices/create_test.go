package invoices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ohheyrj/freeagent/internal/api"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCreateInvoice(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		t.Run(fmt.Sprintf("dry_run=%t", dryRun), func(t *testing.T) {
			var requests []string
			var payload map[string]map[string]any
			client := &api.Client{HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requests = append(requests, r.Method+" "+r.URL.Path)
				body := `{"project":{"contact":"https://api.freeagent.com/v2/contacts/12","currency":"GBP"}}`
				if r.Method == http.MethodPost {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Fatal(err)
					}
					body = `{"invoice":{"url":"https://api.freeagent.com/v2/invoices/99","reference":"056","status":"Draft","total_value":"1200.00","currency":"GBP"}}`
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			cmd := newCreateCommand(func() (*api.Client, error) { return client, nil })
			var out bytes.Buffer
			cmd.SetOut(&out)
			args := []string{"--project=42", "--date=2026-11-01", "--json"}
			if dryRun {
				args = append(args, "--dry-run")
			}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if dryRun {
				if len(requests) != 1 {
					t.Fatalf("dry run made requests: %v", requests)
				}
				if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
			} else {
				if len(requests) != 2 || requests[1] != "POST /v2/invoices" {
					t.Fatalf("unexpected requests: %v", requests)
				}
				var result map[string]any
				if err := json.Unmarshal(out.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if result["id"] != "99" || result["status"] != "Draft" {
					t.Fatalf("unexpected output: %s", out.String())
				}
			}
			invoice := payload["invoice"]
			for key, want := range map[string]any{
				"contact":  "https://api.freeagent.com/v2/contacts/12",
				"project":  strings.TrimRight(api.BaseURL, "/") + "/projects/42",
				"currency": "GBP", "dated_on": "2026-11-01", "payment_terms_in_days": float64(30),
				"include_timeslips": "billed_grouped_by_single_timeslip", "send_new_invoice_emails": false,
				"include_expenses": nil, "include_estimates": nil,
			} {
				got, exists := invoice[key]
				if !exists || got != want {
					t.Errorf("%s = %v (present %t), want %v", key, got, exists, want)
				}
			}
			for _, key := range []string{"reference", "comments", "status"} {
				if _, exists := invoice[key]; exists {
					t.Errorf("unexpected %s in request", key)
				}
			}
		})
	}
}

func TestCreateValidationBeforeClient(t *testing.T) {
	for _, args := range [][]string{
		{}, {"--project=abc"}, {"--project=0"}, {"--project=42", "--date=2026-02-30"},
		{"--project=42", "--payment-terms=-1"}, {"--project=42", "--include-timeslips=invalid"},
		{"--project=42", "--currency=pounds"}, {"--project=42", "unexpected"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := newCreateCommand(func() (*api.Client, error) {
				t.Fatal("invalid input attempted authentication")
				return nil, nil
			})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
