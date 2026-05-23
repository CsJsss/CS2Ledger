package orm

import "github.com/CsJsss/CS2Ledger/pkg/model"

func (o *ormImpl) CreateAccount(a *model.Account) error {
	return o.db.Create(a).Error
}

func (o *ormImpl) ListAccounts() ([]model.Account, error) {
	var accounts []model.Account
	err := o.db.Order("created_at DESC").Find(&accounts).Error
	return accounts, err
}

func (o *ormImpl) FindAccountByID(id uint) (*model.Account, error) {
	var a model.Account
	err := o.db.First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (o *ormImpl) UpdateAccount(a *model.Account) error {
	return o.db.Save(a).Error
}

func (o *ormImpl) DeleteAccount(id uint) error {
	return o.db.Delete(&model.Account{}, id).Error
}

func (o *ormImpl) UpdateAccountInfo(id uint, name string, cookie string) error {
	updates := map[string]any{"name": name}
	if cookie != "" {
		updates["cookie"] = cookie
	}
	return o.db.Model(&model.Account{}).Where("id = ?", id).Updates(updates).Error
}

func (o *ormImpl) UpdateAccountStatus(id uint, status string) error {
	return o.db.Model(&model.Account{}).Where("id = ?", id).
		Update("status", status).Error
}

func (o *ormImpl) UpdateAccountBalanceAndSyncTime(id uint, available, frozen, instant, purchase int64, syncAt int64) error {
	return o.db.Model(&model.Account{}).Where("id = ?", id).Updates(map[string]any{
		"available_balance": available,
		"frozen_balance":    frozen,
		"instant_balance":   instant,
		"purchase_balance":  purchase,
		"last_sync_at":      syncAt,
	}).Error
}
