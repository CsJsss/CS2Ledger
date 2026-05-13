package orm

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

type MigrationFS interface {
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

type ormImpl struct {
	db *gorm.DB
}

func NewORM(sqlDB *sql.DB, migrations MigrationFS) (ORMInterface, error) {
	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("orm: gorm open: %w", err)
	}

	if err := runMigrations(gormDB, migrations); err != nil {
		return nil, fmt.Errorf("orm: migrate: %w", err)
	}

	return &ormImpl{db: gormDB}, nil
}

func NewTestORM() (ORMInterface, error) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("orm test: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("orm: gorm open: %w", err)
	}

	if err := gormDB.AutoMigrate(
		&model.Account{},
		&model.TradeRecord{},
		&model.InventoryItem{},
		&model.RentalRecord{},
		&model.PnlDaily{},
	); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("orm: automigrate: %w", err)
	}

	// AutoMigrate can't create composite unique indexes spanning parent and embedded struct fields.
	// These indexes are defined in migrations/001_initial.sql for the production path.
	if err := gormDB.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_asset ON inventory(account_id, asset_id)",
	).Error; err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("orm: create index: %w", err)
	}

	return &ormImpl{db: gormDB}, nil
}

func runMigrations(db *gorm.DB, migrations MigrationFS) error {
	var currentVersion int
	if db.Migrator().HasTable("schema_version") {
		row := db.Raw("SELECT COALESCE(MAX(version), 0) FROM schema_version").Row()
		if err := row.Scan(&currentVersion); err != nil {
			return err
		}
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		var version int
		if _, err := fmt.Sscanf(f, "%d_", &version); err != nil {
			continue
		}
		if version <= currentVersion {
			continue
		}

		content, err := migrations.ReadFile("migrations/" + f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, stmt := range strings.Split(string(content), ";") {
				stmt = strings.TrimSpace(stmt)
				if stmt == "" || strings.HasPrefix(stmt, "--") {
					continue
				}
				if err := tx.Exec(stmt).Error; err != nil {
					return fmt.Errorf("%s: %w", f, err)
				}
			}
			return tx.Exec(
				"INSERT INTO schema_version (version, applied_at) VALUES (?, unixepoch())",
				version,
			).Error
		}); err != nil {
			return err
		}
	}

	return nil
}
