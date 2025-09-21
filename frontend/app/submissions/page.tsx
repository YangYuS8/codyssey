"use client";
import React from 'react';
import Link from 'next/link';
import {useSubmissions} from '../../src/hooks/useSubmissions';
import {truncate} from '../../src/lib/utils';

export default function SubmissionsPage() {
  const {data, isLoading, error} = useSubmissions({});

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-xl font-semibold">Submissions</h1>
      {isLoading && <div>加载中...</div>}
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
              {data && data.length === 0 && (
                <tr><td colSpan={6} className="py-6 text-center text-gray-400">暂无提交</td></tr>
              )}
              {data?.map(s => (
                <tr key={s.id} className="border-b hover:bg-gray-50">
                  <td className="py-2 pr-4">
                    <Link href={`/submissions/${s.id}`} className="text-blue-600 hover:underline">{truncate(s.id, 12)}</Link>
                  </td>
                  <td className="py-2 pr-4">
                    <Link href={`/problems/${s.problemId}`} className="text-blue-600 hover:underline">{truncate(s.problemId, 12)}</Link>
                  </td>
                  <td className="py-2 pr-4">{s.status}</td>
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
