package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:       "balance <platform>",
	Short:     "Fetch account balance",
	Args:      cobra.ExactArgs(1),
	ValidArgs: validPlatforms,
	RunE:      runBalance,
}

func runBalance(cmd *cobra.Command, args []string) error {
	platformName := args[0]
	token, err := resolveToken(cmd, platformName)
	if err != nil {
		return err
	}
	client, err := createClient(platformName, token)
	if err != nil {
		return err
	}

	b, err := client.GetBalance(context.Background())
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "AVAILABLE\tFROZEN\tINSTANT\tPURCHASE")
	_, _ = fmt.Fprintf(w, "%.2f\t%.2f\t%.2f\t%.2f\n", b.Available, b.Frozen, b.Instant, b.Purchase)
	_ = w.Flush()
	return nil
}
