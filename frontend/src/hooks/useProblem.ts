import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';
import {ProblemSchema, safeParseOrThrow, Problem} from '../api/schemas';

type ProblemDetail = Problem;

async function fetchProblem(id: string): Promise<ProblemDetail> {
  const raw = await apiGet(`/problems/${id}`);
  return safeParseOrThrow(ProblemSchema, raw);
}

export function useProblem(id: string | undefined) {
  return useQuery<ProblemDetail, ApiError>({
    queryKey: ['problem', id],
    queryFn: () => fetchProblem(id!),
    enabled: !!id,
    retry: (failureCount, error) => {
      if (error?.httpStatus === 404) return false;
      return failureCount < 2;
    }
  });
}
