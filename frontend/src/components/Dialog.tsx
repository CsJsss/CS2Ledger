import { ReactNode } from 'react';
import { Dialog as MuiDialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  actions?: ReactNode;
}

export default function Dialog({ open, onClose, title, children, actions }: DialogProps) {
  return (
    <MuiDialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>{children}</DialogContent>
      {actions && <DialogActions>{actions}</DialogActions>}
    </MuiDialog>
  );
}
