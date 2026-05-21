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

var billhistoryCmd = &cobra.Command{
	Use:       "billhistory <platform>",
	Short:     "Fetch bill / fund flow history from a platform",
	Args:      cobra.ExactArgs(1),
	ValidArgs: validPlatforms,
	RunE:      runBillHistory,
}

func init() {
	billhistoryCmd.Flags().Int("limit", 10, "Max records to show")
	billhistoryCmd.Flags().Int64("since", 0, "Unix millisecond timestamp (0 = all)")
	billhistoryCmd.Flags().Bool("raw", false, "Output raw JSON")
}

func runBillHistory(cmd *cobra.Command, args []string) error {
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

	opts := []platform.QueryOption{platform.WithSince(since), platform.WithLimit(limit)}

	bills, err := client.GetBillHistory(context.Background(), opts...)
	if err != nil {
		return err
	}
	if raw {
		printBillsRaw(bills)
	} else {
		printBillsTable(bills)
	}
	return nil
}

func printBillsTable(bills []platform.BillRecord) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "TYPE\tTYPE ID\tMONEY\tORDER NO\tTIME")
	for _, b := range bills {
		_, _ = fmt.Fprintf(w, "%s\t%d\t%.2f\t%s\t%s\n",
			b.TypeName,
			b.TypeID,
			float64(b.ThisMoney)/100.0,
			b.OrderNo,
			time.UnixMilli(b.AddTime).Format(time.DateTime),
		)
	}
	_ = w.Flush()
}
