"use client";
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Button } from '@/src/components/ui/button';

export default function Home() {
  const router = useRouter();
  useEffect(() => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('access_token') : null;
    if (token) {
      // 已登录直接跳到题目列表
      router.replace('/problems');
    }
  }, [router]);

  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-8 text-center gap-8">
      <div className="space-y-4 max-w-xl">
        <h1 className="text-3xl font-bold tracking-tight">Codyssey 在线判题平台</h1>
        <p className="text-neutral-600 text-sm leading-relaxed">
          欢迎！这是项目的前端入口。登录后可浏览题目、创建代码提交并查看判题结果。
          若当前尚未拥有账号，可联系管理员创建，或先浏览公开题目列表（若后端支持匿名访问）。
        </p>
      </div>
      <div className="flex flex-wrap gap-4 items-center justify-center">
        <Link href="/login">
          <Button size="lg">登录</Button>
        </Link>
        <Link href="/problems">
          <Button size="lg" variant="outline">浏览题目</Button>
        </Link>
        <Link href="/submissions">
          <Button size="lg" variant="ghost">查看提交</Button>
        </Link>
      </div>
      <div className="pt-8 text-xs text-neutral-400">© {new Date().getFullYear()} Codyssey</div>
    </main>
  );
}
