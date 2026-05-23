export const PLATFORM_BUFF = 'buff';
export const PLATFORM_YOUPIN = 'youpin';
export const PLATFORM_C5 = 'c5';
export const PLATFORM_IGXE = 'igxe';
export const PLATFORM_ECO = 'eco';
export const PLATFORM_CSQAQ = 'csqaq';

export const PLATFORM_OPTIONS = [
  { value: PLATFORM_BUFF, label: 'BUFF' },
  { value: PLATFORM_YOUPIN, label: '悠悠有品' },
  { value: PLATFORM_C5, label: 'C5' },
  { value: PLATFORM_IGXE, label: 'IGXE' },
  { value: PLATFORM_ECO, label: 'ECO' },
  { value: PLATFORM_CSQAQ, label: 'CSQAQ' },
] as const;

export const priceSourceLabel: Record<string, string> = {
  buff: 'BUFF',
  youpin: '悠悠有品',
  steam: 'Steam',
};

export const platformLabel: Record<string, string> = {
  [PLATFORM_BUFF]: 'BUFF',
  [PLATFORM_YOUPIN]: '悠悠有品',
  [PLATFORM_C5]: 'C5',
  [PLATFORM_IGXE]: 'IGXE',
  [PLATFORM_ECO]: 'ECO',
  [PLATFORM_CSQAQ]: 'CSQAQ',
};

export const inventoryStatusLabel: Record<string, string> = {
  in_inventory: 'In Storage',
  listed: 'Listed',
  rented: 'Rented',
};

export const inventoryStatusColor: Record<string, 'default' | 'success' | 'warning'> = {
  in_inventory: 'default',
  listed: 'success',
  rented: 'warning',
};
