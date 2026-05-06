package orm

import "github.com/CsJsss/CS2Ledger/pkg/model"

func (o *ormImpl) CreateRental(rec *model.RentalRecord) error {
	if rec.ExternalID == "" {
		return o.db.Create(rec).Error
	}
	return o.db.Where("account_id = ? AND external_id = ?", rec.AccountID, rec.ExternalID).
		Assign(rec).FirstOrCreate(rec).Error
}

func (o *ormImpl) FindRentalsByAssetID(accountID uint, assetID string) ([]model.RentalRecord, error) {
	var records []model.RentalRecord
	err := o.db.Where("account_id = ? AND asset_id = ?", accountID, assetID).
		Order("start_at DESC").Find(&records).Error
	return records, err
}

func (o *ormImpl) FindRentalsByAccount(accountID uint) ([]model.RentalRecord, error) {
	var records []model.RentalRecord
	err := o.db.Where("account_id = ?", accountID).
		Order("start_at DESC").Find(&records).Error
	return records, err
}
