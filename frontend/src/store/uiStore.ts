import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ThemeMode = 'light' | 'dark';

interface UIStore {
  selectedAccountId: number | null;
  setSelectedAccount: (id: number | null) => void;
  themeMode: ThemeMode;
  toggleThemeMode: () => void;
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
}

export const useUIStore = create<UIStore>()(
  persist(
    (set) => ({
      selectedAccountId: null,
      setSelectedAccount: (id) => set({ selectedAccountId: id }),
      themeMode: 'dark',
      toggleThemeMode: () => set((s) => ({ themeMode: s.themeMode === 'dark' ? 'light' : 'dark' })),
      sidebarCollapsed: false,
      toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
    }),
    {
      name: 'cs2-ledger-ui',
      partialize: (s) => ({ themeMode: s.themeMode, sidebarCollapsed: s.sidebarCollapsed }),
    },
  ),
);
