import { create } from 'zustand';

interface UIStore {
  selectedAccountId: number | null;
  setSelectedAccount: (id: number | null) => void;
}

export const useUIStore = create<UIStore>((set) => ({
  selectedAccountId: null,
  setSelectedAccount: (id) => set({ selectedAccountId: id }),
}));
