package orm

import "github.com/CsJsss/CS2Ledger/pkg/model"

func (o *ormImpl) CreateBill(r *model.BillRecord) error {
	return o.db.Create(r).Error
}

func (o *ormImpl) ListBillsByAccount(accountID uint, limit, offset int) ([]model.BillRecord, error) {
	var records []model.BillRecord
	err := o.db.Where("account_id = ?", accountID).
		Order("add_time DESC").Limit(limit).Offset(offset).
		Find(&records).Error
	return records, err
}

func (o *ormImpl) ListAllBills(limit, offset int) ([]model.BillRecord, error) {
	var records []model.BillRecord
	err := o.db.Order("add_time DESC").Limit(limit).Offset(offset).
		Find(&records).Error
	return records, err
}

func (o *ormImpl) CountBillsByAccount(accountID uint) (int64, error) {
	var count int64
	err := o.db.Model(&model.BillRecord{}).Where("account_id = ?", accountID).Count(&count).Error
	return count, err
}

func (o *ormImpl) CountAllBills() (int64, error) {
	var count int64
	err := o.db.Model(&model.BillRecord{}).Count(&count).Error
	return count, err
}
