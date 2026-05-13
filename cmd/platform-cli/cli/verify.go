package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:       "verify <platform>",
	Short:     "Verify platform credentials",
	Args:      cobra.ExactArgs(1),
	ValidArgs: validPlatforms,
	RunE:      runVerify,
}

func runVerify(cmd *cobra.Command, args []string) error {
	platformName := args[0]
	token, err := resolveToken(cmd, platformName)
	if err != nil {
		return err
	}
	client, err := createClient(platformName, token)
	if err != nil {
		return err
	}

	if err := client.Verify(context.Background()); err != nil {
		return err
	}
	fmt.Println("OK")
	return nil
}
