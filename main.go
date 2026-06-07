package main

import (
	"context"
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/factory"
	"github.com/CsJsss/CS2Ledger/pkg/service"
	"github.com/CsJsss/CS2Ledger/pkg/service/account"
	"github.com/CsJsss/CS2Ledger/pkg/service/bill"
	"github.com/CsJsss/CS2Ledger/pkg/service/inventory"
	"github.com/CsJsss/CS2Ledger/pkg/service/market"
	"github.com/CsJsss/CS2Ledger/pkg/service/pnl"
	"github.com/CsJsss/CS2Ledger/pkg/service/rental"
	"github.com/CsJsss/CS2Ledger/pkg/service/sync"
	"github.com/CsJsss/CS2Ledger/pkg/service/trade"
	"github.com/CsJsss/CS2Ledger/pkg/utils/configfx"
	"github.com/CsJsss/CS2Ledger/pkg/utils/dbfx"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed migrations/*.sql
var migrationsFS embed.FS
var globalApp *App

func main() {
	container := fx.New(
		configfx.Module,
		fx.Provide(func() orm.MigrationFS { return migrationsFS }),
		dbfx.Module,
		logfx.Module,
		orm.Module,
		platform.Module,
		factory.Module,

		account.Module,
		trade.Module,
		rental.Module,
		bill.Module,
		pnl.Module,
		inventory.Module,
		market.Module,
		sync.Module,
		service.Module,

		logfx.WithComponent("app"),
		fx.Provide(NewApp),
		// Populate fills globalApp with the *App constructed by NewApp,
		// so OnStartup/OnShutdown and Bind can reference it after container.Start.
		fx.Populate(&globalApp),
	)

	ctx := context.Background()
	if err := container.Start(ctx); err != nil {
		os.Exit(1)
	}
	defer func() { _ = container.Stop(ctx) }()

	err := wails.Run(&options.App{
		Title:  "CS2 Ledger",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  globalApp.startup,
		OnShutdown: globalApp.shutdown,
		Bind: []interface{}{
			globalApp,
		},
	})
	if err != nil {
		globalApp.log.Error("wails run failed", "err", err)
	}
}
