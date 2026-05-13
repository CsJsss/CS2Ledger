package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/spf13/cobra"
)

var buyhistoryCmd = &cobra.Command{
	Use:       "buyhistory <platform>",
	Short:     "Fetch buy history from a platform",
	Args:      cobra.ExactArgs(1),
	ValidArgs: validPlatforms,
	RunE:      runBuyHistory,
}

func init() {
	buyhistoryCmd.Flags().Int("limit", 10, "Max records to show")
	buyhistoryCmd.Flags().Int64("since", 0, "Unix millisecond timestamp (0 = all)")
	buyhistoryCmd.Flags().Bool("raw", false, "Output raw JSON")
	buyhistoryCmd.Flags().StringArrayP("query", "Q", nil, "Extra HTTP query param (key=value, repeatable)")
}

func runBuyHistory(cmd *cobra.Command, args []string) error {
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

	trades, err := client.GetBuyHistory(context.Background(), opts...)
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

func printTradesTable(trades []platform.TradeRecord) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ITEM\tEXT\tTRADE TYPE\tQTY\tUNIT PRICE\tTOTAL PRICE\tFEE\tTRADE AT")
	for _, t := range trades {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%.2f\t%.2f\t%.2f\t%s\n",
			t.ItemName,
			t.Exterior,
			t.TradeType,
			t.Quantity,
			float64(t.UnitPrice)/100.0,
			float64(t.TotalPrice)/100.0,
			float64(t.Fee)/100.0,
			time.UnixMilli(t.TradeAt).Format(time.DateTime),
		)
	}
	_ = w.Flush()
}
