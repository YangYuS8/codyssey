import React from 'react';
import {useParams, useRouter} from 'next/navigation';
import {useSubmission} from '../../../src/hooks/useSubmission';
import { useSubmissionEvents } from '@/src/hooks/useSubmissionEvents';
import { useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import {Button} from '../../../src/components/ui/button';
import {truncate} from '../../../src/lib/utils';
import { StatusBadge } from '@/src/components/ui/status-badge';
import { Skeleton } from '@/src/components/ui/skeleton';
import { useRequireAuth } from '@/src/hooks/useRequireAuth';

export default function SubmissionDetailPage() {
  const params = useParams<{id: string}>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const {data, isLoading, error} = useSubmission(params?.id);
  useSubmissionEvents({
    submissionId: params?.id,
    enabled: !!params?.id && !!data && !['ACCEPTED','REJECTED','FAILED','CANCELLED','ERROR','COMPLETED'].includes(data.status.toUpperCase()),
    onEvent: (evt) => {
      if (!params?.id) return;
      if (evt.type === 'status_update' || evt.type === 'judge_run_update' || evt.type === 'completed') {
        // 触发重新获取或乐观更新
        queryClient.invalidateQueries({ queryKey: ['submission', params.id] });
      }
    }
  });
  useRequireAuth();

  if (isLoading) return (
    <div className="p-6 space-y-4 max-w-4xl">
      <div className="flex items-center justify-between">
        <Skeleton className="h-7 w-56" />
        <div className="flex gap-2">
          <Skeleton className="h-8 w-16" />
          <Skeleton className="h-8 w-16" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        {Array.from({ length: 6 }).map((_,i)=>(<Skeleton key={i} className="h-5 w-40" />))}
      </div>
      <Skeleton className="h-6 w-28" />
      <Skeleton className="h-64 w-full" />
    </div>
  );
  if (error) {
    if (error.notFound) return <div className="p-6">提交不存在。</div>;
    return <div className="p-6 text-red-600">加载失败: {error.message}</div>;
  }
  if (!data) return null;

  return (
    <div className="p-6 space-y-4 max-w-4xl">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Submission {truncate(data.id, 18)}</h1>
        <div className="space-x-2">
          <Link href={`/problems/${data.problemId}`}><Button size="sm" variant="outline">题目</Button></Link>
          <Button size="sm" onClick={() => router.back()}>返回</Button>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4 text-sm">
  <div><span className="text-gray-500 mr-1">状态:</span><StatusBadge status={data.status} /></div>
        <div><span className="text-gray-500 mr-1">分数:</span>{data.score ?? '-'}</div>
        <div><span className="text-gray-500 mr-1">语言:</span>{data.language ?? '-'}</div>
        <div><span className="text-gray-500 mr-1">创建:</span>{new Date(data.createdAt).toLocaleString()}</div>
        {data.updatedAt && <div><span className="text-gray-500 mr-1">更新:</span>{new Date(data.updatedAt).toLocaleString()}</div>}
        {data.runtimeMs && <div><span className="text-gray-500 mr-1">耗时(ms):</span>{data.runtimeMs}</div>}
        {data.memoryKb && <div><span className="text-gray-500 mr-1">内存(KB):</span>{data.memoryKb}</div>}
      </div>
      <div className="space-y-2">
        <h2 className="text-sm font-medium">代码</h2>
        <pre className="p-3 bg-neutral-900 text-neutral-100 rounded text-xs overflow-auto max-h-96 whitespace-pre" style={{lineHeight: '1.3'}}>{data.code || '// 无代码内容'}</pre>
      </div>
      {data.judgeRuns && data.judgeRuns.length > 0 && (
        <div className="space-y-2">
          <h2 className="text-sm font-medium">判题执行</h2>
          <table className="min-w-full text-xs">
            <thead>
              <tr className="text-left border-b">
                <th className="py-1 pr-4">ID</th>
                <th className="py-1 pr-4">状态</th>
                <th className="py-1 pr-4">耗时(ms)</th>
                <th className="py-1 pr-4">时间</th>
              </tr>
            </thead>
            <tbody>
              {data.judgeRuns.map(j => (
                <tr key={j.id} className="border-b">
                  <td className="py-1 pr-4">{truncate(j.id, 12)}</td>
                  <td className="py-1 pr-4"><StatusBadge status={j.status} /></td>
                  <td className="py-1 pr-4">{j.durationMs ?? '-'}</td>
                  <td className="py-1 pr-4">{new Date(j.createdAt).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
