import {useMutation, useQueryClient} from '@tanstack/react-query';
import {apiPost, ApiError} from '../api/client';
import type {SubmissionItem} from './useSubmissions';

export interface CreateSubmissionInput {
  problemId: string;
  language: string;
  code: string;
}

async function createSubmission(input: CreateSubmissionInput): Promise<SubmissionItem> {
  return apiPost('/submissions', input);
}

export function useCreateSubmission() {
  const qc = useQueryClient();
  return useMutation<SubmissionItem, ApiError, CreateSubmissionInput>({
    mutationFn: (input) => createSubmission(input),
    onSuccess: (data, variables) => {
      // 失效相关列表缓存
      qc.invalidateQueries({queryKey: ['submissions']});
      qc.invalidateQueries({queryKey: ['submissions', {problemId: variables.problemId}]});
    }
  });
}
