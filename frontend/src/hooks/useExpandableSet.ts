import { useState, useCallback } from 'react';

/**
 * Shared hook for expand/collapse state using a Set of string keys.
 * Used by collapsible day cards in daily sell/buy views.
 */
export function useExpandableSet() {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const toggle = useCallback((key: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }, []);

  const isExpanded = useCallback((key: string) => expanded.has(key), [expanded]);

  return { expanded, isExpanded, toggle };
}
