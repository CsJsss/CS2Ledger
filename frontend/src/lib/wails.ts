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
  GetPnlSummary,
  GetMonthlyBreakdown,
  GetDashboardSummary,
  GetRentalHistory,
} from "../../wailsjs/go/main/App";
import type { model, trade, pnl, sync, inventory, main } from "../../wailsjs/go/models";

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
  GetPnlSummary,
  GetMonthlyBreakdown,
  GetDashboardSummary,
  GetRentalHistory,
};

export type { model, trade, pnl, sync, inventory, main };
