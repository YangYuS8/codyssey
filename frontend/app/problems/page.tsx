"use client";
import React from 'react';
import { Button } from '@/src/components/ui/button';
import { useProblems } from '@/src/hooks/useProblems';
import { Spinner } from '@/src/components/ui/spinner';

export default function ProblemsPage() {
  const { data, isLoading, error } = useProblems();
  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">题目列表</h1>
        <Button size="sm" variant="outline" disabled>新建（待实现）</Button>
      </div>
      <div className="border rounded-md p-4">
        {isLoading && (
          <div className="flex items-center gap-2 text-sm"><Spinner /> <span>加载中...</span></div>
        )}
        {error && (
          <div className="text-sm text-red-600">加载失败：{(error as any)?.message || '未知错误'}</div>
        )}
        {!isLoading && !error && (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left border-b">
                <th className="py-2 pr-4">标题</th>
                <th className="py-2 pr-4 w-40">创建时间</th>
              </tr>
            </thead>
            <tbody>
              {data?.data?.length ? data.data.map(p => (
                <tr key={p.id} className="border-b last:border-b-0 hover:bg-neutral-50 dark:hover:bg-neutral-800/50">
                  <td className="py-2 pr-4 font-medium">{p.title}</td>
                  <td className="py-2 pr-4 text-neutral-500">{new Date(p.created_at).toLocaleString()}</td>
                </tr>
              )) : (
                <tr>
                  <td className="py-4 text-neutral-500" colSpan={2}>暂无数据</td>
                </tr>
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
