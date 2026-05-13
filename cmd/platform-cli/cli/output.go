package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
)

func printTradesRaw(trades []platform.TradeRecord) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(trades); err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
	}
}
