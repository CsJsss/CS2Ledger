package orm

import (
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

func (o *ormImpl) UpsertDailyPnl(accountID uint, tradeAt int64, grossPl, fee, netPl int64) error {
	date := time.UnixMilli(tradeAt).UTC().Format("2006-01-02")
	now := time.Now()

	return o.db.Exec(`
		INSERT INTO pnl_daily (account_id, date, trade_count, gross_pl, fee, net_pl, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?, ?, ?, ?)
		ON CONFLICT(account_id, date) DO UPDATE SET
			trade_count = pnl_daily.trade_count + 1,
			gross_pl = pnl_daily.gross_pl + ?,
			fee = pnl_daily.fee + ?,
			net_pl = pnl_daily.net_pl + ?,
			updated_at = ?
	`, accountID, date, grossPl, fee, netPl, now, now,
		grossPl, fee, netPl, now).Error
}

func (o *ormImpl) FindPnlByAccount(accountID uint) ([]model.PnlDaily, error) {
	var records []model.PnlDaily
	err := o.db.Where("account_id = ?", accountID).
		Order("date DESC").Find(&records).Error
	return records, err
}
