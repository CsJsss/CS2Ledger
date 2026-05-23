import { useState, useEffect } from "react";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import Alert from "@mui/material/Alert";
import { GetSettings, UpdateSettings } from "../lib/wails";

const priceSourceLabels: Record<string, string> = {
  buff: "BUFF",
  youpin: "悠悠有品",
  steam: "Steam",
};

export default function SettingsPage() {
  const [priceSource, setPriceSource] = useState("buff");
  const [cacheTTL, setCacheTTL] = useState(30);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    GetSettings().then((s: { priceSource: string; priceCacheTtl: number }) => {
      if (s) {
        setPriceSource(s.priceSource);
        setCacheTTL(s.priceCacheTtl);
      }
    }).catch(() => {});
  }, []);

  const handleSave = async () => {
    try {
      await UpdateSettings({ priceSource, priceCacheTtl: cacheTTL });
      setSaved(true);
      setError(null);
      setTimeout(() => setSaved(false), 2000);
    } catch (e) {
      setError(String(e));
    }
  };

  return (
    <Box>
      <Typography variant="h4" gutterBottom>设置</Typography>

      <Typography variant="h6" sx={{ mt: 3, mb: 2 }}>行情数据</Typography>

      <Box sx={{ display: "flex", flexDirection: "column", gap: 2, maxWidth: 400 }}>
        <FormControl size="small" fullWidth>
          <InputLabel>默认价格来源</InputLabel>
          <Select
            value={priceSource}
            label="默认价格来源"
            onChange={(e) => setPriceSource(e.target.value)}
          >
            {Object.entries(priceSourceLabels).map(([k, v]) => (
              <MenuItem key={k} value={k}>{v}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <TextField
          label="缓存时间（分钟）"
          type="number"
          value={cacheTTL}
          onChange={(e) => setCacheTTL(Number(e.target.value))}
          inputProps={{ min: 5, max: 1440 }}
          size="small"
          fullWidth
          helperText="范围 5-1440 分钟（最长 24 小时）"
        />

        <Button variant="contained" onClick={() => { void handleSave(); }} sx={{ alignSelf: "flex-start" }}>
          保存
        </Button>

        {saved && <Alert severity="success">设置已保存</Alert>}
        {error && <Alert severity="error">{error}</Alert>}
      </Box>
    </Box>
  );
}
