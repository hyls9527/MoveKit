/**
 * Filter criteria for advanced image filtering.
 * All fields are optional — only set fields are applied.
 */
export interface FilterCriteria {
  /** Filter by model name (exact match from metadata_json) */
  model?: string;
  /** Minimum rating (0-5) */
  ratingMin?: number;
  /** Maximum rating (0-5) */
  ratingMax?: number;
  /** Only show favorites */
  favorite?: boolean;
  /** Filter by format (png, jpg, webp, avif) */
  format?: string;
  /** Start date (ISO string, inclusive) */
  dateFrom?: string;
  /** End date (ISO string, inclusive) */
  dateTo?: string;
}

/**
 * Check if a FilterCriteria has any active filters.
 */
export function hasActiveFilters(criteria: FilterCriteria): boolean {
  return (
    criteria.model !== undefined ||
    criteria.ratingMin !== undefined ||
    criteria.ratingMax !== undefined ||
    criteria.favorite !== undefined ||
    criteria.format !== undefined ||
    criteria.dateFrom !== undefined ||
    criteria.dateTo !== undefined
  );
}

/**
 * Count the number of active filters.
 */
export function countActiveFilters(criteria: FilterCriteria): number {
  let count = 0;
  if (criteria.model !== undefined) count++;
  if (criteria.ratingMin !== undefined || criteria.ratingMax !== undefined) count++;
  if (criteria.favorite !== undefined) count++;
  if (criteria.format !== undefined) count++;
  if (criteria.dateFrom !== undefined || criteria.dateTo !== undefined) count++;
  return count;
}
