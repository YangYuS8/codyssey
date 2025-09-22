import React from 'react';
import {useProblem} from '../../../src/hooks/useProblem';
import {useParams, useRouter} from 'next/navigation';
import Link from 'next/link';
import {Button} from '../../../src/components/ui/button';
import { Skeleton } from '@/src/components/ui/skeleton';
import { useRequireAuth } from '@/src/hooks/useRequireAuth';

export default function ProblemDetailPage() {
  const params = useParams<{id: string}>();
  const router = useRouter();
  const {data, isLoading, error} = useProblem(params?.id);
  useRequireAuth();

  if (isLoading) return (
    <div className="p-6 space-y-4 max-w-3xl">
      <Skeleton className="h-8 w-72" />
      <Skeleton className="h-5 w-24" />
      <div className="flex gap-2">
        {Array.from({ length: 4 }).map((_,i)=>(<Skeleton key={i} className="h-6 w-16" />))}
      </div>
      <Skeleton className="h-64 w-full" />
    </div>
  );
  if (error) {
    if (error.notFound) return <div className="p-6">题目不存在。</div>;
    return <div className="p-6 text-red-600">加载失败: {error.message}</div>;
  }
  if (!data) return null;

  return (
    <div className="p-6 space-y-4 max-w-3xl">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">{data.title}</h1>
        <div className="space-x-2">
          <Link href={`/problems/${data.id}/submit`}><Button size="sm">提交代码</Button></Link>
          <Button variant="outline" size="sm" onClick={() => router.back()}>返回</Button>
        </div>
      </div>
      {data.difficulty && <div className="text-sm text-gray-500">难度: {data.difficulty}</div>}
      {data.tags && data.tags.length > 0 && (
        <div className="flex flex-wrap gap-2 text-xs text-gray-600">
          {data.tags.map(t => <span key={t} className="bg-gray-100 px-2 py-1 rounded">{t}</span>)}
        </div>
      )}
      <article className="prose prose-sm max-w-none whitespace-pre-wrap leading-relaxed">
        {data.description || '（暂无描述）'}
      </article>
    </div>
  );
}
