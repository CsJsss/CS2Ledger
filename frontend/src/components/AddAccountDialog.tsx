import { useState, useEffect } from "react";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Box from "@mui/material/Box";
import Dialog from "./Dialog";
import { PLATFORM_OPTIONS } from "../lib/constants";

interface AddAccountDialogProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; platform: string; cookie: string }) => void;
  isPending: boolean;
  error: string | null;
  editMode?: boolean;
  initialValues?: { name: string; platform: string };
}

export default function AddAccountDialog({
  open,
  onClose,
  onSubmit,
  isPending,
  error,
  editMode = false,
  initialValues,
}: AddAccountDialogProps) {
  const [name, setName] = useState("");
  const [platform, setPlatform] = useState<string>(PLATFORM_OPTIONS[0].value);
  const [cookie, setCookie] = useState("");

  useEffect(() => {
    if (editMode && initialValues) {
      setName(initialValues.name);
      setPlatform(initialValues.platform);
      setCookie("");
    } else if (!editMode) {
      setName("");
      setPlatform(PLATFORM_OPTIONS[0].value);
      setCookie("");
    }
  }, [open, editMode, initialValues]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    if (!editMode && !cookie.trim()) return;
    onSubmit({ name: name.trim(), platform, cookie: cookie.trim() });
  };

  const handleCancel = () => {
    onClose();
    if (!editMode) {
      setName("");
      setCookie("");
    }
  };

  const formId = editMode ? "edit-account-form" : "add-account-form";

  return (
    <Dialog
      open={open}
      onClose={handleCancel}
      title={editMode ? "Edit Account" : "Add Account"}
      actions={
        <>
          <Button onClick={handleCancel}>Cancel</Button>
          <Button type="submit" variant="contained" disabled={isPending} form={formId}>
            {isPending ? "Saving..." : editMode ? "Save Changes" : "Add"}
          </Button>
        </>
      }
    >
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      <Box component="form" id={formId} onSubmit={handleSubmit} sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
        <TextField
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="My BUFF account"
          required
          size="small"
          fullWidth
        />

        <FormControl size="small" fullWidth>
          <InputLabel>Platform</InputLabel>
          <Select
            value={platform}
            label="Platform"
            onChange={(e) => setPlatform(e.target.value)}
            disabled={editMode}
          >
            {PLATFORM_OPTIONS.map((p) => (
              <MenuItem key={p.value} value={p.value}>{p.label}</MenuItem>
            ))}
          </Select>
        </FormControl>

        <TextField
          label="Cookie"
          value={cookie}
          onChange={(e) => setCookie(e.target.value)}
          placeholder={editMode ? "Leave empty to keep current cookie" : "Paste your platform cookie here..."}
          required={!editMode}
          multiline
          rows={4}
          size="small"
          fullWidth
        />
      </Box>
    </Dialog>
  );
}
