import TextField from "@mui/material/TextField";
import InputAdornment from "@mui/material/InputAdornment";
import SearchIcon from "@mui/icons-material/Search";

interface PageSearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export default function PageSearchBar({ value, onChange, placeholder = "Search..." }: PageSearchBarProps) {
  return (
    <TextField
      size="small"
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      InputProps={{
        startAdornment: (
          <InputAdornment position="start">
            <SearchIcon fontSize="small" color="action" />
          </InputAdornment>
        ),
      }}
      sx={{
        width: 260,
        "& .MuiOutlinedInput-root": { bgcolor: "grey.100" },
      }}
    />
  );
}
