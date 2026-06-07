package orm

import (
	"gorm.io/gorm"

	"github.com/CsJsss/CS2Ledger/pkg/model"
)

type BillFilter struct {
	TypeID    int
	Platform  string
	StartTime int64 // unix ms, 0 = no filter
	EndTime   int64 // unix ms, 0 = no filter
}

func applyBillFilter(q *gorm.DB, f BillFilter) *gorm.DB {
	if f.TypeID != 0 {
		q = q.Where("type_id = ?", f.TypeID)
	}
	if f.Platform != "" {
		q = q.Where("platform = ?", f.Platform)
	}
	if f.StartTime > 0 {
		q = q.Where("add_time >= ?", f.StartTime)
	}
	if f.EndTime > 0 {
		q = q.Where("add_time <= ?", f.EndTime)
	}
	return q
}

func (o *ormImpl) CreateBill(r *model.BillRecord) error {
	return o.db.Create(r).Error
}

func (o *ormImpl) ListBillsByAccount(accountID uint, limit, offset int, f BillFilter) ([]model.BillRecord, error) {
	var records []model.BillRecord
	q := o.db.Where("account_id = ?", accountID)
	q = applyBillFilter(q, f)
	err := q.Order("add_time DESC").Limit(limit).Offset(offset).Find(&records).Error
	return records, err
}

func (o *ormImpl) ListAllBills(limit, offset int, f BillFilter) ([]model.BillRecord, error) {
	var records []model.BillRecord
	q := o.db.
		Joins("JOIN accounts ON accounts.id = bill_records.account_id").
		Where("accounts.deleted_at IS NULL")
	q = applyBillFilter(q, f)
	err := q.Order("add_time DESC").Limit(limit).Offset(offset).Find(&records).Error
	return records, err
}

func (o *ormImpl) CountBillsByAccount(accountID uint, f BillFilter) (int64, error) {
	var count int64
	q := o.db.Model(&model.BillRecord{}).Where("account_id = ?", accountID)
	q = applyBillFilter(q, f)
	err := q.Count(&count).Error
	return count, err
}

func (o *ormImpl) CountAllBills(f BillFilter) (int64, error) {
	var count int64
	q := o.db.Model(&model.BillRecord{}).
		Joins("JOIN accounts ON accounts.id = bill_records.account_id").
		Where("accounts.deleted_at IS NULL")
	q = applyBillFilter(q, f)
	err := q.Count(&count).Error
	return count, err
}

// DailyBillSummary is a single day's aggregated bill totals (in fen).
type DailyBillSummary struct {
	Date      string // "2006-01-02"
	TypeID    int
	ThisMoney int64
}

// SumBillByDay returns daily total this_money grouped by type_id, for chart rendering.
func (o *ormImpl) SumBillByDay(accountID uint, f BillFilter) ([]DailyBillSummary, error) {
	var rows []DailyBillSummary
	q := o.db.Model(&model.BillRecord{}).
		Select("DATE(ROUND(add_time / 1000), 'unixepoch') AS date, type_id, SUM(this_money) AS this_money")
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	}
	q = applyBillFilter(q, f)
	err := q.Group("date, type_id").Order("date ASC").Scan(&rows).Error
	return rows, err
}

func (o *ormImpl) SumBillsByTypes(accountID uint, typeIDs []int) (int64, error) {
	var sum int64
	q := o.db.Model(&model.BillRecord{})
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	}
	if len(typeIDs) > 0 {
		q = q.Where("type_id IN ?", typeIDs)
	}
	err := q.Select("COALESCE(SUM(this_money), 0)").Scan(&sum).Error
	return sum, err
}
