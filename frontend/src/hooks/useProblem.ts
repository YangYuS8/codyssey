import {useQuery} from '@tanstack/react-query';
import {apiGet, ApiError} from '../api/client';

export interface ProblemDetail {
  id: string;
  title: string;
  description?: string;
  difficulty?: string;
  tags?: string[];
  createdAt?: string;
  updatedAt?: string;
}

async function fetchProblem(id: string): Promise<ProblemDetail> {
  return apiGet(`/problems/${id}`);
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
