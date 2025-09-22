"use client";
import React, { useState, useEffect } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useSubmissions } from '../../src/hooks/useSubmissions';
import { truncate } from '../../src/lib/utils';
import { Pagination } from '@/src/components/ui/pagination';
import { StatusBadge } from '@/src/components/ui/status-badge';
import { useRequireAuth } from '@/src/hooks/useRequireAuth';
import { Skeleton } from '@/src/components/ui/skeleton';

export default function SubmissionsPage() {
  const sp = useSearchParams();
  const router = useRouter();
  const [page, setPage] = useState<number>(() => Number(sp.get('page') || 1));
  const [pageSize] = useState(15);
  const [status, setStatus] = useState(() => sp.get('status') || '');
  const [language, setLanguage] = useState(() => sp.get('language') || '');
  const [problemId, setProblemId] = useState(() => sp.get('problemId') || '');
  const {data, isLoading, error} = useSubmissions({ page, pageSize, status: status || undefined, language: language || undefined, problemId: problemId || undefined });
  const total = data?.meta.filtered || 0;
  useRequireAuth();

  useEffect(() => {
    const params = new URLSearchParams();
    if (page && page !== 1) params.set('page', String(page));
    if (status) params.set('status', status);
    if (language) params.set('language', language);
    if (problemId) params.set('problemId', problemId);
    const qs = params.toString();
    const target = qs ? `/submissions?${qs}` : '/submissions';
    router.replace(target, { scroll: false });
  }, [page, status, language, problemId, router]);

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-xl font-semibold">Submissions</h1>
      <div className="flex flex-wrap gap-4 items-end">
        <div className="space-y-1">
          <label className="text-xs text-neutral-500">题目ID</label>
          <input value={problemId} onChange={e => { setPage(1); setProblemId(e.target.value); }} className="border rounded px-2 py-1 text-sm" placeholder="精确匹配" />
        </div>
        <div className="space-y-1">
          <label className="text-xs text-neutral-500">状态</label>
          <select value={status} onChange={e => { setPage(1); setStatus(e.target.value); }} className="border rounded px-2 py-1 text-sm bg-white">
            <option value="">全部</option>
            <option value="pending">Pending</option>
            <option value="running">Running</option>
            <option value="accepted">Accepted</option>
            <option value="rejected">Rejected</option>
            <option value="failed">Failed</option>
          </select>
        </div>
        <div className="space-y-1">
          <label className="text-xs text-neutral-500">语言</label>
          <select value={language} onChange={e => { setPage(1); setLanguage(e.target.value); }} className="border rounded px-2 py-1 text-sm bg-white">
            <option value="">全部</option>
            <option value="cpp">C++</option>
            <option value="python">Python</option>
            <option value="go">Go</option>
            <option value="java">Java</option>
          </select>
        </div>
        <Pagination
          page={page}
            pageSize={pageSize}
            total={total}
            onPageChange={setPage}
            className="ml-auto"
        />
      </div>
      {isLoading && (
        <div className="space-y-3 min-h-[250px]">
          <div className="text-sm">加载中...</div>
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-7 w-full" />
          ))}
        </div>
      )}
      {error && <div className="text-red-600 text-sm">加载失败: {error.message}</div>}
      {!isLoading && !error && (
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="text-left border-b">
                <th className="py-2 pr-4">ID</th>
                <th className="py-2 pr-4">题目</th>
                <th className="py-2 pr-4">状态</th>
                <th className="py-2 pr-4">分数</th>
                <th className="py-2 pr-4">语言</th>
                <th className="py-2 pr-4">时间</th>
              </tr>
            </thead>
            <tbody>
              {data && data.data.length === 0 && (
                <tr><td colSpan={6} className="py-6 text-center text-gray-400">暂无提交</td></tr>
              )}
              {data?.data.map(s => (
                <tr key={s.id} className="border-b hover:bg-gray-50">
                  <td className="py-2 pr-4">
                    <Link href={`/submissions/${s.id}`} className="text-blue-600 hover:underline">{truncate(s.id, 12)}</Link>
                  </td>
                  <td className="py-2 pr-4">
                    <Link href={`/problems/${s.problemId}`} className="text-blue-600 hover:underline">{truncate(s.problemId, 12)}</Link>
                  </td>
                  <td className="py-2 pr-4"><StatusBadge status={s.status} /></td>
                  <td className="py-2 pr-4">{s.score ?? '-'}</td>
                  <td className="py-2 pr-4">{s.language ?? '-'}</td>
                  <td className="py-2 pr-4 text-xs text-gray-500">{new Date(s.createdAt).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
