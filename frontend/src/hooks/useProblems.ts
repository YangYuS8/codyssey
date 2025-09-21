"use client";
import { useQuery } from '@tanstack/react-query';
import { apiGet } from '@/src/api/client';
import { ProblemSchema, safeParseOrThrow } from '@/src/api/schemas';

export interface Problem {
  id: string;
  title: string;
  description?: string;
  createdAt?: string;
  difficulty?: string;
  tags?: string[];
}

export interface ProblemListResponse {
  data: Problem[];
  meta: { page: number; pageSize: number; total: number; filtered: number };
}
export interface UseProblemsParams {
  page?: number;
  pageSize?: number;
  search?: string;
  difficulty?: string;
}

export function useProblems(params: UseProblemsParams = {}) {
  const { page = 1, pageSize = 20, search = '', difficulty } = params;
  return useQuery<ProblemListResponse>({
    queryKey: ['problems', { page, pageSize, search, difficulty }],
    queryFn: async () => {
      // 暂时仍请求全部（后端分页对接前），然后在前端进行过滤 + 分页
      const all = await apiGet<any[]>(`/problems`);
      const normalized: Problem[] = all.map(p => {
        const v = safeParseOrThrow(ProblemSchema, p) as any;
        return {
          id: v.id,
          title: v.title,
          description: v.description,
          difficulty: v.difficulty,
          tags: v.tags,
          createdAt: v.createdAt,
        } as Problem;
      });
      let filtered = normalized;
      if (search) {
        const s = search.toLowerCase();
        filtered = filtered.filter(p => p.title.toLowerCase().includes(s));
      }
      if (difficulty) {
        filtered = filtered.filter(p => p.difficulty === difficulty);
      }
      const total = normalized.length;
      const filteredCount = filtered.length;
      const start = (page - 1) * pageSize;
      const end = start + pageSize;
      const pageData = filtered.slice(start, end);
      return {
        data: pageData,
        meta: { page, pageSize, total, filtered: filteredCount }
      };
    },
  });
}
