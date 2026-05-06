package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/service/account"
	"github.com/CsJsss/CS2Ledger/pkg/service/inventory"
	"github.com/CsJsss/CS2Ledger/pkg/service/pnl"
	"github.com/CsJsss/CS2Ledger/pkg/service/trade"
	"github.com/CsJsss/CS2Ledger/pkg/utils/dbfx"
)

const defaultDSN = "file:demo.db"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	db, err := dbfx.NewDB(dbfx.Config{DSN: defaultDSN})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ormInst, err := orm.NewORM(db, osMigrationFS{"migrations"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init ORM: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "accounts":
		listAccounts(account.NewService(nil, ormInst))
	case "create-account":
		if len(os.Args) < 5 {
			fmt.Println("Usage: demo create-account <name> <platform> <cookie>")
			return
		}
		createAccount(account.NewService(nil, ormInst), os.Args[2], os.Args[3], os.Args[4])
	case "inventory":
		listInventory(inventory.NewService(nil, ormInst))
	case "trades":
		listTrades(trade.NewService(nil, ormInst))
	case "pnl":
		showPnl(pnl.NewService(nil, ormInst))
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("CS2 Ledger CLI Demo")
	fmt.Println("Database: " + defaultDSN)
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  demo accounts         List all accounts")
	fmt.Println("  demo create-account   Create a new account")
	fmt.Println("  demo inventory        List inventory items")
	fmt.Println("  demo trades           List completed trades")
	fmt.Println("  demo pnl              Show P&L summary")
}

func listAccounts(svc account.AccountInterface) {
	accounts, err := svc.List()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("%-4s %-20s %-10s %-10s\n", "ID", "Name", "Platform", "Status")
	for _, a := range accounts {
		fmt.Printf("%-4d %-20s %-10s %-10s\n", a.ID, a.Name, a.Platform, a.Status)
	}
}

func createAccount(svc account.AccountInterface, name, platform, cookie string) {
	a, err := svc.Create(name, platform, cookie)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Created: ID=%d Name=%s Platform=%s\n", a.ID, a.Name, a.Platform)
}

func listInventory(svc inventory.InventoryInterface) {
	items, err := svc.List(0, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("%-4s %-30s %-12s\n", "ID", "Item", "Status")
	for _, item := range items {
		fmt.Printf("%-4d %-30s %-12s\n", item.ID, item.ItemName, item.Status)
	}
}

func listTrades(svc trade.TradeInterface) {
	trades, err := svc.ListCompletedTrades(0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("%-4s %-30s %-8s %-8s\n", "ID", "Item", "NetPL", "Fee")
	for _, t := range trades {
		fmt.Printf("%-4d %-30s %-8d %-8d\n", t.SellTradeID, t.ItemName, t.NetPl, t.TotalFee)
	}
}

func showPnl(svc pnl.PnlInterface) {
	summary, err := svc.GetSummary(0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Total Trades: %d\n", summary.TotalTrades)
	fmt.Printf("Gross P&L:    %d fen\n", summary.TotalGrossPl)
	fmt.Printf("Total Fees:   %d fen\n", summary.TotalFee)
	fmt.Printf("Net P&L:      %d fen\n", summary.TotalNetPl)
}

type osMigrationFS struct{ root string }

func (o osMigrationFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(o.root, name))
}

func (o osMigrationFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(o.root, name))
}
