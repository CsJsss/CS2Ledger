import { useState } from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router';
import Box from '@mui/material/Box';
import Drawer from '@mui/material/Drawer';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import Typography from '@mui/material/Typography';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import Chip from '@mui/material/Chip';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import SearchIcon from '@mui/icons-material/Search';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import LightModeIcon from '@mui/icons-material/LightMode';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import MenuIcon from '@mui/icons-material/Menu';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import DashboardIcon from '@mui/icons-material/Dashboard';
import InventoryIcon from '@mui/icons-material/Inventory';
import ReceiptIcon from '@mui/icons-material/Receipt';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import SettingsIcon from '@mui/icons-material/Settings';
import { useAccounts } from '../hooks/useAccounts';
import { useUIStore } from '../store/uiStore';

const DRAWER_WIDTH = 224;
const DRAWER_COLLAPSED = 56;
const APPBAR_HEIGHT = 56;

const navItems = [
  { to: '/dashboard', label: '仪表盘', icon: <DashboardIcon fontSize="small" /> },
  { to: '/inventory', label: '持仓', icon: <InventoryIcon fontSize="small" /> },
  { to: '/trades/completed', label: '交易记录', icon: <ReceiptIcon fontSize="small" /> },
  { to: '/pnl', label: '盈亏', icon: <TrendingUpIcon fontSize="small" /> },
  { to: '/accounts', label: '账户管理', icon: <AccountBalanceIcon fontSize="small" /> },
  { to: '/settings', label: '设置', icon: <SettingsIcon fontSize="small" /> },
];

export default function AppLayout() {
  const { data: accounts = [] } = useAccounts();
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const setSelectedAccount = useUIStore((s) => s.setSelectedAccount);
  const themeMode = useUIStore((s) => s.themeMode);
  const toggleThemeMode = useUIStore((s) => s.toggleThemeMode);
  const sidebarCollapsed = useUIStore((s) => s.sidebarCollapsed);
  const toggleSidebar = useUIStore((s) => s.toggleSidebar);
  const navigate = useNavigate();

  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
  const [searchText, setSearchText] = useState('');

  const selectedAccount = accounts.find((a) => a.ID === selectedAccountId) ?? null;
  const platformLabel = (p: string) =>
    ({ buff: 'BUFF', youpin: '悠悠', c5: 'C5', igxe: 'IGXE', eco: 'ECO' })[p] ?? p;

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && searchText.trim()) {
      void navigate(`/inventory?search=${encodeURIComponent(searchText.trim())}`);
    }
  };

  const drawerWidth = sidebarCollapsed ? DRAWER_COLLAPSED : DRAWER_WIDTH;

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      <AppBar
        position="fixed"
        color="default"
        elevation={1}
        sx={{ zIndex: (t) => t.zIndex.drawer + 1, height: APPBAR_HEIGHT }}
      >
        <Toolbar sx={{ height: APPBAR_HEIGHT, minHeight: `${APPBAR_HEIGHT}px !important`, gap: 2 }}>
          <IconButton size="small" onClick={toggleSidebar} sx={{ color: 'text.secondary' }}>
            {sidebarCollapsed ? (
              <MenuIcon fontSize="small" />
            ) : (
              <ChevronLeftIcon fontSize="small" />
            )}
          </IconButton>

          <Typography variant="h6" fontWeight="bold" sx={{ whiteSpace: 'nowrap' }}>
            CS2 Ledger
          </Typography>

          <TextField
            size="small"
            placeholder="搜索饰品、交易..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onKeyDown={handleSearchKeyDown}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" color="action" />
                </InputAdornment>
              ),
            }}
            sx={{
              maxWidth: 360,
              flex: 1,
              '& .MuiOutlinedInput-root': { bgcolor: 'background.paper' },
            }}
          />

          <Box sx={{ flex: 1 }} />

          <IconButton
            size="small"
            onClick={toggleThemeMode}
            sx={{ color: 'text.secondary' }}
            title={themeMode === 'dark' ? '切换浅色模式' : '切换深色模式'}
          >
            {themeMode === 'dark' ? (
              <LightModeIcon fontSize="small" />
            ) : (
              <DarkModeIcon fontSize="small" />
            )}
          </IconButton>

          <Tooltip
            title={
              selectedAccount
                ? `${selectedAccount.name} (${platformLabel(selectedAccount.platform)})`
                : '所有账号'
            }
          >
            <Chip
              label={
                selectedAccount
                  ? `${selectedAccount.name} · ${platformLabel(selectedAccount.platform)}`
                  : '所有账号'
              }
              variant="outlined"
              size="small"
              deleteIcon={<ExpandMoreIcon />}
              onDelete={() => {}}
              onClick={(e) => setAnchorEl(e.currentTarget)}
              sx={{
                cursor: 'pointer',
                maxWidth: 220,
                bgcolor: selectedAccount ? 'rgba(249,115,22,0.12)' : 'transparent',
                borderColor: selectedAccount ? 'rgba(249,115,22,0.3)' : undefined,
                color: selectedAccount ? '#f97316' : 'text.secondary',
              }}
            />
          </Tooltip>

          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={() => setAnchorEl(null)}
            anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
            transformOrigin={{ vertical: 'top', horizontal: 'right' }}
          >
            <MenuItem
              selected={selectedAccountId === null}
              onClick={() => {
                setSelectedAccount(null);
                setAnchorEl(null);
              }}
            >
              所有账号
            </MenuItem>
            {accounts.map((a) => (
              <MenuItem
                key={a.ID}
                selected={a.ID === selectedAccountId}
                onClick={() => {
                  setSelectedAccount(a.ID);
                  setAnchorEl(null);
                }}
              >
                {a.name} ({platformLabel(a.platform)})
              </MenuItem>
            ))}
          </Menu>
        </Toolbar>
      </AppBar>

      <Box sx={{ display: 'flex', flex: 1, mt: `${APPBAR_HEIGHT}px` }}>
        <Drawer
          variant="permanent"
          sx={{
            width: drawerWidth,
            flexShrink: 0,
            transition: 'width 0.2s ease',
            '& .MuiDrawer-paper': {
              width: drawerWidth,
              boxSizing: 'border-box',
              mt: `${APPBAR_HEIGHT}px`,
              transition: 'width 0.2s ease',
              overflowX: 'hidden',
            },
          }}
        >
          <List sx={{ flex: 1, px: 1, pt: 2 }}>
            {navItems.map((item) => (
              <Tooltip key={item.to} title={sidebarCollapsed ? item.label : ''} placement="right">
                <ListItemButton
                  component={NavLink}
                  to={item.to}
                  sx={{
                    borderRadius: 1,
                    mb: 0.5,
                    gap: 1.5,
                    justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
                    minHeight: 40,
                    px: sidebarCollapsed ? 1 : undefined,
                    '&.active': {
                      bgcolor: 'rgba(249,115,22,0.12)',
                      color: '#f97316',
                      borderLeft: '3px solid #f97316',
                      borderRadius: '0 6px 6px 0',
                      pl: sidebarCollapsed ? 1 : 1.5,
                    },
                    '&.active:hover': { bgcolor: 'rgba(249,115,22,0.18)' },
                    '&:hover': { bgcolor: 'action.hover' },
                  }}
                >
                  <Box
                    component="span"
                    sx={{ color: 'inherit', display: 'flex', alignItems: 'center' }}
                  >
                    {item.icon}
                  </Box>
                  {!sidebarCollapsed && (
                    <ListItemText
                      primary={item.label}
                      primaryTypographyProps={{ fontSize: 13, fontWeight: 500 }}
                    />
                  )}
                </ListItemButton>
              </Tooltip>
            ))}
          </List>
        </Drawer>

        <Box component="main" sx={{ flex: 1, overflow: 'auto', p: 3 }}>
          <Outlet />
        </Box>
      </Box>
    </Box>
  );
}
