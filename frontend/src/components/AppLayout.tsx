import { useState } from "react";
import { Outlet, NavLink, useNavigate } from "react-router";
import Box from "@mui/material/Box";
import Drawer from "@mui/material/Drawer";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";
import Typography from "@mui/material/Typography";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import TextField from "@mui/material/TextField";
import InputAdornment from "@mui/material/InputAdornment";
import Chip from "@mui/material/Chip";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Tooltip from "@mui/material/Tooltip";
import SearchIcon from "@mui/icons-material/Search";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import { useAccounts } from "../hooks/useAccounts";
import { useUIStore } from "../store/uiStore";

const DRAWER_WIDTH = 224;
const APPBAR_HEIGHT = 56;

const navItems = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/inventory", label: "Inventory" },
  { to: "/trades/completed", label: "Trades" },
  { to: "/pnl", label: "P&L" },
  { to: "/accounts", label: "Accounts" },
  { to: "/settings", label: "Settings" },
];

export default function AppLayout() {
  const { data: accounts = [] } = useAccounts();
  const selectedAccountId = useUIStore((s) => s.selectedAccountId);
  const setSelectedAccount = useUIStore((s) => s.setSelectedAccount);
  const navigate = useNavigate();

  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
  const [searchText, setSearchText] = useState("");

  const selectedAccount = accounts.find((a) => a.ID === selectedAccountId) ?? null;
  const platformLabel = (p: string) =>
    ({ buff: "BUFF", youpin: "悠悠", c5: "C5", igxe: "IGXE" }[p] ?? p);

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && searchText.trim()) {
      void navigate(`/inventory?search=${encodeURIComponent(searchText.trim())}`);
    }
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", height: "100vh" }}>
      <AppBar
        position="fixed"
        color="default"
        elevation={1}
        sx={{ zIndex: (t) => t.zIndex.drawer + 1, height: APPBAR_HEIGHT }}
      >
        <Toolbar sx={{ height: APPBAR_HEIGHT, minHeight: `${APPBAR_HEIGHT}px !important`, gap: 2 }}>
          <Typography variant="h6" fontWeight="bold" sx={{ whiteSpace: "nowrap" }}>
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
              "& .MuiOutlinedInput-root": { bgcolor: "grey.100" },
            }}
          />

          <Box sx={{ flex: 1 }} />

          <Tooltip title={selectedAccount ? `${selectedAccount.name} (${platformLabel(selectedAccount.platform)})` : "所有账号"}>
            <Chip
              label={
                selectedAccount
                  ? `${selectedAccount.name} · ${platformLabel(selectedAccount.platform)}`
                  : "所有账号"
              }
              color={selectedAccount ? "primary" : "default"}
              variant={selectedAccount ? "filled" : "outlined"}
              size="small"
              deleteIcon={<ExpandMoreIcon />}
              onDelete={() => {}}
              onClick={(e) => setAnchorEl(e.currentTarget)}
              sx={{ cursor: "pointer", maxWidth: 220 }}
            />
          </Tooltip>

          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={() => setAnchorEl(null)}
            anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
            transformOrigin={{ vertical: "top", horizontal: "right" }}
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

      <Box sx={{ display: "flex", flex: 1, mt: `${APPBAR_HEIGHT}px` }}>
        <Drawer
          variant="permanent"
          sx={{
            width: DRAWER_WIDTH,
            flexShrink: 0,
            "& .MuiDrawer-paper": {
              width: DRAWER_WIDTH,
              boxSizing: "border-box",
              mt: `${APPBAR_HEIGHT}px`,
            },
          }}
        >
          <List sx={{ flex: 1, px: 1, pt: 2 }}>
            {navItems.map((item) => (
              <ListItemButton
                key={item.to}
                component={NavLink}
                to={item.to}
                sx={{
                  borderRadius: 1,
                  mb: 0.5,
                  "&.active": { bgcolor: "primary.main", color: "white" },
                  "&.active:hover": { bgcolor: "primary.dark" },
                }}
              >
                <ListItemText primary={item.label} primaryTypographyProps={{ fontSize: 14, fontWeight: 500 }} />
              </ListItemButton>
            ))}
          </List>
        </Drawer>

        <Box component="main" sx={{ flex: 1, overflow: "auto", p: 3 }}>
          <Outlet />
        </Box>
      </Box>
    </Box>
  );
}
