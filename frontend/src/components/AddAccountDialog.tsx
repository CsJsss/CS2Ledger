import { useState, useEffect } from 'react';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Box from '@mui/material/Box';
import Dialog from './Dialog';
import { PLATFORM_OPTIONS, PLATFORM_CSQAQ } from '../lib/constants';

const RSA_PLATFORMS = new Set(['eco', 'igxe']);

function useResettableState(open: boolean, editMode: boolean, initial: string) {
  const [value, setValue] = useState(initial);
  useEffect(() => {
    setValue(editMode ? '' : initial);
  }, [open, editMode, initial]);
  return [value, setValue] as const;
}

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
  const [name, setName] = useResettableState(open, editMode, '');
  const [platform, setPlatform] = useState<string>(PLATFORM_OPTIONS[0].value);
  const [cookie, setCookie] = useResettableState(open, editMode, '');
  const [identityId, setIdentityId] = useResettableState(open, editMode, '');
  const [rsaKey, setRsaKey] = useResettableState(open, editMode, '');

  useEffect(() => {
    if (editMode && initialValues) {
      setName(initialValues.name);
      setPlatform(initialValues.platform);
    } else if (!editMode) {
      setPlatform(PLATFORM_OPTIONS[0].value);
    }
  }, [open, editMode, initialValues, setName]);

  const isRSA = RSA_PLATFORMS.has(platform);

  const buildCredential = (): string => {
    if (platform === PLATFORM_CSQAQ) return cookie.trim();
    if (isRSA) {
      if (!identityId.trim() && !rsaKey.trim()) return '';
      return identityId.trim() + ':' + rsaKey.trim();
    }
    return cookie.trim();
  };

  const credentialEmpty =
    platform === PLATFORM_CSQAQ
      ? !cookie.trim()
      : isRSA
        ? !identityId.trim() || !rsaKey.trim()
        : !cookie.trim();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    if (!editMode && credentialEmpty) return;
    onSubmit({ name: name.trim(), platform, cookie: buildCredential() });
  };

  const handleCancel = () => {
    onClose();
    if (!editMode) {
      setName('');
      setCookie('');
      setIdentityId('');
      setRsaKey('');
    }
  };

  const formId = editMode ? 'edit-account-form' : 'add-account-form';

  return (
    <Dialog
      open={open}
      onClose={handleCancel}
      title={editMode ? '编辑账户' : '添加账户'}
      actions={
        <>
          <Button onClick={handleCancel}>取消</Button>
          <Button type="submit" variant="contained" disabled={isPending} form={formId}>
            {isPending ? '保存中...' : editMode ? '保存更改' : '添加'}
          </Button>
        </>
      }
    >
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Box
        component="form"
        id={formId}
        onSubmit={handleSubmit}
        sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}
      >
        <TextField
          label="名称"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="例如：我的 BUFF 账户"
          required
          size="small"
          fullWidth
        />

        <FormControl size="small" fullWidth>
          <InputLabel>平台</InputLabel>
          <Select
            value={platform}
            label="平台"
            onChange={(e) => setPlatform(e.target.value)}
            disabled={editMode}
          >
            {PLATFORM_OPTIONS.map((p) => (
              <MenuItem key={p.value} value={p.value}>
                {p.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {platform === PLATFORM_CSQAQ ? (
          <TextField
            label="API Token"
            value={cookie}
            onChange={(e) => setCookie(e.target.value)}
            placeholder={editMode ? '留空则不更改当前 Token' : '在此粘贴 CSQAQ API Token...'}
            required={!editMode}
            size="small"
            fullWidth
          />
        ) : isRSA ? (
          <>
            <TextField
              label="身份ID"
              value={identityId}
              onChange={(e) => setIdentityId(e.target.value)}
              placeholder="合作方 ID"
              required={!editMode}
              size="small"
              fullWidth
            />
            <TextField
              label="RSA 私钥"
              value={rsaKey}
              onChange={(e) => setRsaKey(e.target.value)}
              placeholder={editMode ? '留空则不更改当前密钥' : '粘贴 RSA 私钥 (PEM 格式)...'}
              required={!editMode}
              multiline
              rows={4}
              size="small"
              fullWidth
            />
          </>
        ) : (
          <TextField
            label="Cookie"
            value={cookie}
            onChange={(e) => setCookie(e.target.value)}
            placeholder={editMode ? '留空则不更改当前 Cookie' : '在此粘贴平台 Cookie...'}
            required={!editMode}
            multiline
            rows={4}
            size="small"
            fullWidth
          />
        )}
      </Box>
    </Dialog>
  );
}
