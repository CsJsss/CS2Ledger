package account

import (
	"go.uber.org/fx"

	"github.com/CsJsss/CS2Ledger/pkg/model"
	"github.com/CsJsss/CS2Ledger/pkg/orm"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type AccountInterface interface {
	List() ([]model.Account, error)
	Get(id uint) (*model.Account, error)
	Create(name, platform, cookie string, withdrawalFeeRate int64) (*model.Account, error)
	Update(a *model.Account) error
	UpdateInfo(id uint, name string, cookie string, withdrawalFeeRate int64) error
	Delete(id uint) error
}

type service struct {
	log *logfx.Logger
	orm orm.ORMInterface
}

func NewService(log *logfx.Logger, orm orm.ORMInterface) *service {
	return &service{log: log, orm: orm}
}

func (s *service) List() ([]model.Account, error) {
	return s.orm.ListAccounts()
}

func (s *service) Get(id uint) (*model.Account, error) {
	return s.orm.FindAccountByID(id)
}

func (s *service) Create(name, platform, cookie string, withdrawalFeeRate int64) (*model.Account, error) {
	a := &model.Account{
		Name:              name,
		Platform:          platform,
		Cookie:            cookie,
		WithdrawalFeeRate: withdrawalFeeRate,
		Status:            orm.AccountStatusActive,
	}
	if err := s.orm.CreateAccount(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) Update(a *model.Account) error {
	return s.orm.UpdateAccount(a)
}

func (s *service) UpdateInfo(id uint, name string, cookie string, withdrawalFeeRate int64) error {
	return s.orm.UpdateAccountInfo(id, name, cookie, withdrawalFeeRate)
}

func (s *service) Delete(id uint) error {
	return s.orm.DeleteAccount(id)
}

var Module = fx.Module("account",
	logfx.WithComponent("account"),
	fx.Provide(
		NewService,
		fx.Annotate(func(s *service) AccountInterface { return s }, fx.As(new(AccountInterface))),
	),
)
