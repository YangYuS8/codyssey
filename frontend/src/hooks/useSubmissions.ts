import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';
import {SubmissionItemSchema, safeParseOrThrow, SubmissionItem} from '../api/schemas';

export type { SubmissionItem };

export interface UseSubmissionsParams {
  problemId?: string;
  status?: string;
  page?: number;
  pageSize?: number;
  language?: string;
}

interface SubmissionsListResponse {
  data: SubmissionItem[];
  meta: { page: number; pageSize: number; total: number; filtered: number };
}

async function fetchSubmissions(params: UseSubmissionsParams): Promise<SubmissionsListResponse> {
  // 暂时全量获取再前端过滤
  const qs = new URLSearchParams();
  if (params.problemId) qs.set('problemId', params.problemId);
  const query = qs.toString();
  const raw = await apiGet(`/submissions${query ? `?${query}` : ''}`);
  if (!Array.isArray(raw)) throw new Error('SCHEMA_VALIDATION_FAILED_LIST_TYPE');
  let list = raw.map((r: unknown) => safeParseOrThrow(SubmissionItemSchema, r) as SubmissionItem);
  const total = list.length;
  if (params.status) list = list.filter(s => s.status.toLowerCase() === params.status!.toLowerCase());
  if (params.language) list = list.filter(s => (s.language || '').toLowerCase() === params.language!.toLowerCase());
  const filtered = list.length;
  const page = params.page || 1;
  const pageSize = params.pageSize || 20;
  const start = (page - 1) * pageSize;
  const end = start + pageSize;
  return {
    data: list.slice(start, end),
    meta: { page, pageSize, total, filtered }
  };
}

export function useSubmissions(params: UseSubmissionsParams) {
  return useQuery<SubmissionsListResponse, ApiError>({
    queryKey: ['submissions', params],
    queryFn: () => fetchSubmissions(params),
  });
}
