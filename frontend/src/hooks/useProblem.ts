import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';
import {ProblemSchema, safeParseOrThrow, Problem} from '../api/schemas';

async function fetchProblem(id: string): Promise<Problem> {
  const raw = await apiGet(`/problems/${id}`);
  const parsed = safeParseOrThrow(ProblemSchema, raw) as Problem;
  return parsed;
}

export function useProblem(id: string | undefined) {
  return useQuery<Problem, ApiError>({
    queryKey: ['problem', id],
    queryFn: () => fetchProblem(id!),
    enabled: !!id,
    retry: (failureCount, error) => {
      if (error?.httpStatus === 404) return false;
      return failureCount < 2;
    }
  });
}
