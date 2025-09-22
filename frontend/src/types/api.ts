// 统一类型定义
export interface PageMeta {
  page: number;
  pageSize: number;
  total: number;
  filtered: number;
}

export interface ApiEnvelope<T, M = unknown> {
  data: T;
  meta?: M; // 列表响应可带分页或其他元信息
  error?: { code: string; message: string };
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: PageMeta;
}

export class SchemaValidationError extends Error {
  issues: unknown;
  constructor(issues: unknown) {
    super('SCHEMA_VALIDATION_FAILED');
    this.issues = issues;
  }
}

// Web Vitals 指标结构（Next.js metric 对象）
export interface WebVitalMetric {
  id: string;
  name: string; // FCP / LCP / CLS / INP / TTFB / FID
  label: 'web-vital' | string;
  value: number;
  startTime: number;
  delta?: number;
}
