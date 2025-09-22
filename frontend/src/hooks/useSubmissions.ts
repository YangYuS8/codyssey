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
  const serverPagination = process.env.NEXT_PUBLIC_SERVER_PAGINATION === 'true';
  const { page = 1, pageSize = 20, problemId, status, language } = params;
  if (serverPagination) {
    const q = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
    if (problemId) q.set('problemId', problemId);
    if (status) q.set('status', status);
    if (language) q.set('language', language);
    const resp = await apiGet<{ data: unknown[]; meta?: { page:number; pageSize:number; total:number; filtered:number } }>(`/submissions?${q.toString()}`);
    const items = Array.isArray(resp.data) ? resp.data : [];
  const data: SubmissionItem[] = items.map(r => safeParseOrThrow(SubmissionItemSchema, r) as SubmissionItem);
    const meta = resp.meta || { page, pageSize, total: data.length, filtered: data.length };
    return { data, meta };
  }
  // 前端回退
  const raw = await apiGet<unknown[]>(`/submissions${problemId ? `?problemId=${encodeURIComponent(problemId)}` : ''}`);
  let list: SubmissionItem[] = raw.map(r => safeParseOrThrow(SubmissionItemSchema, r) as SubmissionItem);
  const total = list.length;
  if (status) list = list.filter(s => s.status.toLowerCase() === status.toLowerCase());
  if (language) list = list.filter(s => (s.language || '').toLowerCase() === language.toLowerCase());
  const filtered = list.length;
  const start = (page - 1) * pageSize;
  const end = start + pageSize;
  return { data: list.slice(start, end), meta: { page, pageSize, total, filtered } };
}

export function useSubmissions(params: UseSubmissionsParams) {
  return useQuery<SubmissionsListResponse, ApiError>({
    queryKey: ['submissions', params],
    queryFn: () => fetchSubmissions(params),
  });
}
