import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';
import {SubmissionDetailSchema, safeParseOrThrow, SubmissionDetail as SchemaSubmissionDetail} from '../api/schemas';

export type SubmissionDetail = SchemaSubmissionDetail;

async function fetchSubmission(id: string): Promise<SubmissionDetail> {
  const raw = await apiGet(`/submissions/${id}`);
  return safeParseOrThrow(SubmissionDetailSchema, raw) as SubmissionDetail;
}

export interface UseSubmissionOptions {
  enablePolling?: boolean;
}

export function useSubmission(id: string | undefined, opts: UseSubmissionOptions = {}) {
  const { enablePolling = true } = opts;
  return useQuery<SubmissionDetail, ApiError>({
    queryKey: ['submission', id],
    queryFn: () => fetchSubmission(id!),
    enabled: !!id,
    refetchInterval: enablePolling ? (query) => {
      const d = query.state.data as SubmissionDetail | undefined;
      if (!d) return 4000;
      const terminal = ['ACCEPTED','REJECTED','FAILED','CANCELLED','ERROR','COMPLETED'];
      return terminal.includes(d.status.toUpperCase()) ? false : 4000;
    } : false,
    retry: (count, err) => {
      if (err?.httpStatus === 404) return false;
      return count < 2;
    }
  });
}
