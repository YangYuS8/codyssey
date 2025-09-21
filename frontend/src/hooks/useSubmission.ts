import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';
import type {SubmissionItem} from './useSubmissions';

export interface SubmissionDetail extends SubmissionItem {
  code?: string;
  runtimeMs?: number;
  memoryKb?: number;
  judgeRuns?: Array<{
    id: string;
    status: string;
    durationMs?: number;
    createdAt: string;
  }>;
}

async function fetchSubmission(id: string): Promise<SubmissionDetail> {
  return apiGet(`/submissions/${id}`);
}

export function useSubmission(id: string | undefined) {
  return useQuery<SubmissionDetail, ApiError>({
    queryKey: ['submission', id],
    queryFn: () => fetchSubmission(id!),
    enabled: !!id,
    retry: (count, err) => {
      if (err?.httpStatus === 404) return false;
      return count < 2;
    }
  });
}
