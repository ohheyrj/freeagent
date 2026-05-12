package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ohheyrj/freeagent/internal/api"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Run OAuth login to obtain access tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")

		clientID := os.Getenv("FREEAGENT_CLIENT_ID")
		clientSecret := os.Getenv("FREEAGENT_CLIENT_SECRET")

		if err := api.Login(clientID, clientSecret, port); err != nil {
			return err
		}
		fmt.Println("Authentication successful! Tokens saved.")
		return nil
	},
}

func init() {
	loginCmd.Flags().Int("port", 8080, "local port for OAuth callback")
	Cmd.AddCommand(loginCmd)
}
