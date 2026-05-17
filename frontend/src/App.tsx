import { Routes, Route, Navigate } from 'react-router';
import AppLayout from './components/AppLayout';
import DashboardPage from './pages/DashboardPage';
import InventoryPage from './pages/InventoryPage';
import InventoryDetailPage from './pages/InventoryDetailPage';
import CompletedTradesPage from './pages/CompletedTradesPage';
import PnLPage from './pages/PnLPage';
import BillPage from './pages/BillPage';
import AccountsPage from './pages/AccountsPage';
import SettingsPage from './pages/SettingsPage';

export default function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="inventory" element={<InventoryPage />} />
        <Route path="inventory/:accountId/:assetId" element={<InventoryDetailPage />} />
        <Route path="trades" element={<Navigate to="/trades/completed" replace />} />
        <Route path="trades/completed" element={<CompletedTradesPage />} />
        <Route path="pnl" element={<PnLPage />} />
        <Route path="bill" element={<BillPage />} />
        <Route path="accounts" element={<AccountsPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>
    </Routes>
  );
}
