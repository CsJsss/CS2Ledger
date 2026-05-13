package cli

import (
	"context"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/spf13/cobra"
)

var sellhistoryCmd = &cobra.Command{
	Use:       "sellhistory <platform>",
	Short:     "Fetch sell history from a platform",
	Args:      cobra.ExactArgs(1),
	ValidArgs: validPlatforms,
	RunE:      runSellHistory,
}

func init() {
	sellhistoryCmd.Flags().Int("limit", 10, "Max records to show")
	sellhistoryCmd.Flags().Int64("since", 0, "Unix millisecond timestamp (0 = all)")
	sellhistoryCmd.Flags().Bool("raw", false, "Output raw JSON")
	sellhistoryCmd.Flags().StringArrayP("query", "Q", nil, "Extra HTTP query param (key=value, repeatable)")
}

func runSellHistory(cmd *cobra.Command, args []string) error {
	platformName := args[0]
	token, err := resolveToken(cmd, platformName)
	if err != nil {
		return err
	}
	client, err := createClient(platformName, token)
	if err != nil {
		return err
	}

	limit, _ := cmd.Flags().GetInt("limit")
	since, _ := cmd.Flags().GetInt64("since")
	raw, _ := cmd.Flags().GetBool("raw")
	extra, _ := cmd.Flags().GetStringArray("query")

	opts := []platform.QueryOption{platform.WithSince(since), platform.WithLimit(limit)}
	if params := parseExtraParams(extra); len(params) > 0 {
		opts = append(opts, platform.WithExtraParams(params))
	}

	trades, err := client.GetSellHistory(context.Background(), opts...)
	if err != nil {
		return err
	}
	if raw {
		printTradesRaw(trades)
	} else {
		printTradesTable(trades)
	}
	return nil
}
