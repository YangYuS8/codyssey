"use client";
import React, { useState, useEffect } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { Button } from '@/src/components/ui/button';
import { Pagination } from '@/src/components/ui/pagination';
import { useProblems } from '@/src/hooks/useProblems';
import { Spinner } from '@/src/components/ui/spinner';
import { useRequireAuth } from '@/src/hooks/useRequireAuth';
import { Skeleton } from '@/src/components/ui/skeleton';

export default function ProblemsPage() {
  const sp = useSearchParams();
  const router = useRouter();
  const [page, setPage] = useState<number>(() => Number(sp.get('page') || 1));
  const [pageSize] = useState(10);
  const [search, setSearch] = useState(() => sp.get('search') || '');
  const [difficulty, setDifficulty] = useState<string | undefined>(() => sp.get('difficulty') || undefined);
  const { data, isLoading, error } = useProblems({ page, pageSize, search, difficulty });
  useRequireAuth();
  const total = data?.meta.filtered || 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  // 同步到 URL（防抖可选，这里简单即时同步）
  useEffect(() => {
    const params = new URLSearchParams();
    if (page && page !== 1) params.set('page', String(page));
    if (search) params.set('search', search);
    if (difficulty) params.set('difficulty', difficulty);
    const qs = params.toString();
    const target = qs ? `/problems?${qs}` : '/problems';
    router.replace(target, { scroll: false });
  }, [page, search, difficulty, router]);

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">题目列表</h1>
        <Button size="sm" variant="outline" disabled>新建（待实现）</Button>
      </div>
      <div className="flex flex-wrap gap-4 items-end">
        <div className="space-y-1">
          <label className="text-xs text-neutral-500">搜索</label>
          <input
            value={search}
            onChange={e => { setPage(1); setSearch(e.target.value); }}
            placeholder="按标题搜索"
            className="border rounded px-2 py-1 text-sm"
          />
        </div>
        <div className="space-y-1">
          <label className="text-xs text-neutral-500">难度</label>
          <select
            className="border rounded px-2 py-1 text-sm bg-white"
            value={difficulty || ''}
            onChange={e => { setPage(1); setDifficulty(e.target.value || undefined); }}
          >
            <option value="">全部</option>
            <option value="easy">Easy</option>
            <option value="medium">Medium</option>
            <option value="hard">Hard</option>
          </select>
        </div>
        <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage} className="ml-auto" />
      </div>
      <div className="border rounded-md p-4 min-h-[300px]">
        {isLoading && (
          <div className="space-y-3">
            <div className="flex items-center gap-2 text-sm"><Spinner /> <span>加载中...</span></div>
            <Skeleton className="h-5 w-3/5" />
            <Skeleton className="h-5 w-2/5" />
            <Skeleton className="h-5 w-4/5" />
            <div className="space-y-2 pt-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-8 w-full" />
              ))}
            </div>
          </div>
        )}
        {error && (
          <div className="text-sm text-red-600">加载失败：{(error as any)?.message || '未知错误'}</div>
        )}
        {!isLoading && !error && (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left border-b">
                <th className="py-2 pr-4">标题</th>
                <th className="py-2 pr-4 w-32">难度</th>
                <th className="py-2 pr-4 w-40">创建时间</th>
              </tr>
            </thead>
            <tbody>
              {data?.data?.length ? data.data.map(p => (
                <tr key={p.id} className="border-b last:border-b-0 hover:bg-neutral-50 dark:hover:bg-neutral-800/50">
                  <td className="py-2 pr-4 font-medium">{p.title}</td>
                  <td className="py-2 pr-4 text-neutral-500">{p.difficulty || '-'}</td>
                  <td className="py-2 pr-4 text-neutral-500">{p.createdAt ? new Date(p.createdAt).toLocaleString() : '-'}</td>
                </tr>
              )) : (
                <tr>
                  <td className="py-4 text-neutral-500" colSpan={3}>暂无数据</td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
