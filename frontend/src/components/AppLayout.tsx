import { Outlet, NavLink } from "react-router";
import Box from "@mui/material/Box";
import Drawer from "@mui/material/Drawer";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";
import Typography from "@mui/material/Typography";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import FormControl from "@mui/material/FormControl";
import { useAccounts } from "../hooks/useAccounts";
import { useUIStore } from "../store/uiStore";

const DRAWER_WIDTH = 224;

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

  return (
    <Box sx={{ display: "flex", height: "100vh" }}>
      <Drawer
        variant="permanent"
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          "& .MuiDrawer-paper": { width: DRAWER_WIDTH, boxSizing: "border-box" },
        }}
      >
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" fontWeight="bold">CS2 Ledger</Typography>
        </Box>

        <List sx={{ flex: 1, px: 1 }}>
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

        {accounts.length > 1 && (
          <Box sx={{ p: 2, borderTop: 1, borderColor: "divider" }}>
            <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, display: "block" }}>
              Account
            </Typography>
            <FormControl size="small" fullWidth>
              <Select
                value={selectedAccountId ?? ""}
                onChange={(e) => setSelectedAccount(e.target.value ? Number(e.target.value) : null)}
                displayEmpty
              >
                <MenuItem value="">All accounts</MenuItem>
                {accounts.map((a) => (
                  <MenuItem key={a.ID} value={a.ID}>
                    {a.name} ({a.platform})
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>
        )}
      </Drawer>

      <Box component="main" sx={{ flex: 1, overflow: "auto", p: 3 }}>
        <Outlet />
      </Box>
    </Box>
  );
}
