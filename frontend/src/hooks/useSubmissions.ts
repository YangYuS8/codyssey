import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';

export interface SubmissionItem {
  id: string;
  problemId: string;
  userId: string;
  status: string;
  score?: number;
  language?: string;
  createdAt: string;
  updatedAt?: string;
  version?: number;
}

export interface SubmissionsListParams {
  problemId?: string;
}

async function fetchSubmissions(params: SubmissionsListParams): Promise<SubmissionItem[]> {
  const qs = new URLSearchParams();
  if (params.problemId) qs.set('problemId', params.problemId);
  const query = qs.toString();
  return apiGet(`/submissions${query ? `?${query}` : ''}`);
}

export function useSubmissions(params: SubmissionsListParams) {
  return useQuery<SubmissionItem[], ApiError>({
    queryKey: ['submissions', params],
    queryFn: () => fetchSubmissions(params),
    staleTime: 5_000,
  });
}
