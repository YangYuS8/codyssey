"use client";
import { useQuery } from '@tanstack/react-query';
import { apiGet } from '@/src/api/client';

export interface Problem {
  id: string;
  title: string;
  description: string;
  created_at: string;
}

interface ProblemListResponse {
  data: Problem[]; // 服务端 envelope 解包后实际直接返回 data 数组，这里保留结构以便后续扩展 meta
  meta?: { limit: number; offset: number; count: number; total?: number };
}

export function useProblems(params: { limit?: number; offset?: number } = {}) {
  const { limit = 20, offset = 0 } = params;
  return useQuery({
    queryKey: ['problems', { limit, offset }],
    queryFn: async () => {
      const qs = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
      // apiGet 直接返回 data 部分，因此需要一个包装（后端 envelope: { data: [..], meta, error }）
      // 当前实现 apiGet<T> 返回 data，所以我们需要再请求 meta 的话要新建一个方法，这里暂时只取列表
      const list = await apiGet<Problem[]>(`/problems?${qs}`);
      return { data: list } as ProblemListResponse;
    },
  });
}
