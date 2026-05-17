package platform

import (
	"context"
	"fmt"
)

// FakeClient implements Client with configurable responses for testing.
type FakeClient struct {
	VerifyErr   error
	BuyHistory  []TradeRecord
	SellHistory []TradeRecord
	BalanceResp *Balance
	BillHistory []BillRecord
}

func (f *FakeClient) Verify(_ context.Context) error {
	if f.VerifyErr != nil {
		return fmt.Errorf("fake verify: %w", f.VerifyErr)
	}
	return nil
}

func (f *FakeClient) GetBuyHistory(_ context.Context, _ ...QueryOption) ([]TradeRecord, error) {
	return f.BuyHistory, nil
}

func (f *FakeClient) GetSellHistory(_ context.Context, _ ...QueryOption) ([]TradeRecord, error) {
	return f.SellHistory, nil
}

func (f *FakeClient) GetBalance(_ context.Context) (*Balance, error) {
	if f.BalanceResp != nil {
		return f.BalanceResp, nil
	}
	return &Balance{}, nil
}

func (f *FakeClient) GetBillHistory(_ context.Context, _ ...QueryOption) ([]BillRecord, error) {
	return f.BillHistory, nil
}
