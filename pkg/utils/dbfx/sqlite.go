package dbfx

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/utils/configfx"
)

var Module = fx.Module("dbfx",
	fx.Provide(func(cfg configfx.Config) Config {
		return Config{
			DSN:         cfg.DB.DSN,
			WALMode:     cfg.DB.WALMode,
			BusyTimeout: cfg.DB.BusyTimeout,
		}
	}),
	fx.Provide(NewDB),
)

type Config struct {
	DSN         string
	WALMode     bool
	BusyTimeout int
}

func NewDB(cfg Config) (*sql.DB, error) {
	if cfg.BusyTimeout == 0 {
		cfg.BusyTimeout = 5000
	}

	dsn := cfg.DSN
	params := make([]string, 0, 2)
	if cfg.WALMode {
		params = append(params, "_journal_mode=WAL")
	}
	params = append(params, fmt.Sprintf("_busy_timeout=%d", cfg.BusyTimeout))

	for i, p := range params {
		if i == 0 {
			dsn += "?" + p
		} else {
			dsn += "&" + p
		}
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}

	return db, nil
}
