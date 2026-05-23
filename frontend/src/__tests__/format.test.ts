import { describe, it, expect } from 'vitest';
import { formatCNY, plColor, fmt, plHexColor } from '../lib/format';

describe('formatCNY', () => {
  it('formats positive cents as CNY', () => {
    expect(formatCNY(12345)).toBe('¥123.45');
  });

  it('formats zero', () => {
    expect(formatCNY(0)).toBe('¥0.00');
  });

  it('formats negative values', () => {
    expect(formatCNY(-1050)).toBe('-¥10.50');
  });

  it('returns "¥ --" for null', () => {
    expect(formatCNY(null)).toBe('¥ --');
  });

  it('returns "¥ --" for undefined', () => {
    expect(formatCNY(undefined)).toBe('¥ --');
  });
});

describe('plColor', () => {
  it('returns "success" for positive values', () => {
    expect(plColor(100)).toBe('success');
  });

  it('returns "error" for negative values', () => {
    expect(plColor(-100)).toBe('error');
  });

  it('returns "text.secondary" for zero', () => {
    expect(plColor(0)).toBe('text.secondary');
  });
});

describe('fmt', () => {
  it('formats a number to 2 decimal places', () => {
    expect(fmt(12.5)).toBe('12.50');
  });

  it('returns 0.00 for null', () => {
    expect(fmt(null)).toBe('0.00');
  });

  it('returns 0.00 for undefined', () => {
    expect(fmt(undefined)).toBe('0.00');
  });
});

describe('plHexColor', () => {
  it('returns green for positive values', () => {
    expect(plHexColor(100)).toBe('#16a34a');
  });

  it('returns red for negative values', () => {
    expect(plHexColor(-100)).toBe('#dc2626');
  });

  it('returns grey for zero', () => {
    expect(plHexColor(0)).toBe('#6b7280');
  });
});
