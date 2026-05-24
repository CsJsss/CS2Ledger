import { createTheme, type Theme } from '@mui/material/styles';

declare module '@mui/material/styles' {
  interface Palette {
    accent: Palette['primary'];
    profit: Palette['primary'];
    loss: Palette['primary'];
  }
  interface PaletteOptions {
    accent?: PaletteOptions['primary'];
    profit?: PaletteOptions['primary'];
    loss?: PaletteOptions['primary'];
  }
}

const typography = {
  fontFamily: '"Geist Variable", sans-serif',
  h4: { fontSize: '1.75rem', fontWeight: 700, letterSpacing: '-0.02em' },
  h5: { fontSize: '1.25rem', fontWeight: 600 },
  h6: { fontSize: '1.1rem', fontWeight: 600 },
  body2: { fontSize: '0.875rem' },
  caption: { fontSize: '0.75rem' },
} as const;

const shape = { borderRadius: 8 } as const;

const sharedPalette = {
  primary: { main: '#f97316', dark: '#ea580c', contrastText: '#ffffff' },
  error: { main: '#ef4444' },
  success: { main: '#22c55e' },
  warning: { main: '#f59e0b' },
  accent: { main: '#f97316', dark: '#ea580c', contrastText: '#ffffff' },
  profit: { main: '#22c55e' },
  loss: { main: '#ef4444' },
} as const;

const sharedComponents = {
  MuiCssBaseline: {
    styleOverrides: { body: { margin: 0 } },
  },
  MuiButton: {
    styleOverrides: {
      root: { textTransform: 'none', fontWeight: 500 },
      contained: { boxShadow: 'none' },
    },
  },
  MuiPaper: {
    styleOverrides: { root: { backgroundImage: 'none' } },
  },
  MuiListItemButton: {
    styleOverrides: {
      root: {
        borderRadius: 6,
        marginBottom: 2,
        '&.active': {
          backgroundColor: 'rgba(249,115,22,0.12)',
          color: '#f97316',
          borderLeft: '3px solid #f97316',
        },
        '&.active:hover': { backgroundColor: 'rgba(249,115,22,0.18)' },
      },
    },
  },
  MuiTableRow: {
    styleOverrides: {
      head: { '&:hover': { backgroundColor: 'transparent !important' } },
    },
  },
  MuiTableCell: {
    styleOverrides: {
      head: {
        fontWeight: 500,
        fontSize: '0.7rem',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
      },
    },
  },
  MuiChip: {
    styleOverrides: { root: { fontWeight: 500, fontSize: '0.75rem' } },
  },
  MuiTextField: {
    styleOverrides: {
      root: {
        '& .MuiOutlinedInput-root': {
          '&.Mui-focused fieldset': { borderColor: '#f97316' },
        },
      },
    },
  },
  MuiSelect: {
    styleOverrides: {
      root: {
        '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: '#f97316' },
      },
    },
  },
  MuiTabs: {
    styleOverrides: { indicator: { backgroundColor: '#f97316' } },
  },
  MuiMenuItem: {
    styleOverrides: {
      root: {
        '&.Mui-selected': { backgroundColor: 'rgba(249,115,22,0.12)' },
      },
    },
  },
  MuiAlert: {
    styleOverrides: { root: { borderRadius: 8 } },
  },
} as const;

// eslint-disable-next-line @typescript-eslint/no-unsafe-argument
const darkTheme = createTheme({
  palette: {
    mode: 'dark',
    ...sharedPalette,
    secondary: { main: '#a1a1aa' },
    background: { default: '#09090b', paper: '#18181b' },
  },
  typography,
  shape,
  components: {
    ...sharedComponents,
    MuiCssBaseline: {
      styleOverrides: { body: { backgroundColor: '#09090b' } },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          backgroundImage: 'linear-gradient(135deg, #18181b, #1f1f23)',
          border: '1px solid rgba(255,255,255,0.08)',
          boxShadow: 'none',
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: '#0d0d10',
          borderBottom: '1px solid rgba(255,255,255,0.08)',
          boxShadow: 'none',
        },
        colorDefault: { backgroundColor: '#0d0d10' },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          backgroundColor: '#0d0d10',
          borderRight: '1px solid rgba(255,255,255,0.08)',
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: { borderRadius: 8 },
        standardError: {
          backgroundColor: 'rgba(239,68,68,0.1)',
          border: '1px solid rgba(239,68,68,0.2)',
        },
        standardSuccess: {
          backgroundColor: 'rgba(34,197,94,0.1)',
          border: '1px solid rgba(34,197,94,0.2)',
        },
        standardWarning: {
          backgroundColor: 'rgba(245,158,11,0.1)',
          border: '1px solid rgba(245,158,11,0.2)',
        },
      },
    },
  },
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
} as any);

// eslint-disable-next-line @typescript-eslint/no-unsafe-argument
const lightTheme = createTheme({
  palette: {
    mode: 'light',
    ...sharedPalette,
    secondary: { main: '#64748b' },
    background: { default: '#f8fafc', paper: '#ffffff' },
  },
  typography,
  shape,
  components: {
    ...sharedComponents,
    MuiCard: {
      styleOverrides: {
        root: {
          border: '1px solid #e2e8f0',
          boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        standardError: {
          backgroundColor: '#fef2f2',
          border: '1px solid #fecaca',
        },
        standardSuccess: {
          backgroundColor: '#f0fdf4',
          border: '1px solid #bbf7d0',
        },
        standardWarning: {
          backgroundColor: '#fffbeb',
          border: '1px solid #fde68a',
        },
      },
    },
  },
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
} as any);

const themes: Record<'light' | 'dark', Theme> = { light: lightTheme, dark: darkTheme };

export function getTheme(mode: 'light' | 'dark'): Theme {
  return themes[mode];
}

export default darkTheme;
