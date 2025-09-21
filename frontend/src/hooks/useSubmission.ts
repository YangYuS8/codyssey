import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';
import type {SubmissionItem} from './useSubmissions';
import {SubmissionDetailSchema, safeParseOrThrow, SubmissionDetail as SchemaSubmissionDetail} from '../api/schemas';

export interface SubmissionDetail extends SchemaSubmissionDetail {}

async function fetchSubmission(id: string): Promise<SubmissionDetail> {
  const raw = await apiGet(`/submissions/${id}`);
  return safeParseOrThrow(SubmissionDetailSchema, raw) as SubmissionDetail;
}

export function useSubmission(id: string | undefined) {
  return useQuery<SubmissionDetail, ApiError>({
    queryKey: ['submission', id],
    queryFn: () => fetchSubmission(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const d = query.state.data as SubmissionDetail | undefined;
      if (!d) return 4000;
      const terminal = ['ACCEPTED','REJECTED','FAILED','CANCELLED','ERROR','COMPLETED'];
      return terminal.includes(d.status.toUpperCase()) ? false : 4000;
    },
    retry: (count, err) => {
      if (err?.httpStatus === 404) return false;
      return count < 2;
    }
  });
}
