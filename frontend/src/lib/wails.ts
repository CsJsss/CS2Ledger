import {
  GetAccounts,
  CreateAccount,
  UpdateAccount,
  UpdateAccountInfo,
  DeleteAccount,
  SyncAccount,
  GetInventory,
  GetItemDetail,
  GetCompletedTrades,
  GetCompletedTradesSummary,
  GetUnmatchedSells,
  GetPnlSummary,
  GetMonthlyBreakdown,
  GetDashboardSummary,
  GetRentalHistory,
  GetMarketPrices,
  GetSettings,
  UpdateSettings,
} from '../../wailsjs/go/main/App';
import type { model, trade, pnl, sync, inventory, main } from '../../wailsjs/go/models';

export {
  GetAccounts,
  CreateAccount,
  UpdateAccount,
  UpdateAccountInfo,
  DeleteAccount,
  SyncAccount,
  GetInventory,
  GetItemDetail,
  GetCompletedTrades,
  GetCompletedTradesSummary,
  GetUnmatchedSells,
  GetPnlSummary,
  GetMonthlyBreakdown,
  GetDashboardSummary,
  GetRentalHistory,
  GetMarketPrices,
  GetSettings,
  UpdateSettings,
};

export type { model, trade, pnl, sync, inventory, main };
