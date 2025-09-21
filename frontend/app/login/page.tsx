"use client";
import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/src/auth/auth-context';
import { Button } from '@/src/components/ui/button';
import { Input } from '@/src/components/ui/input';
import { Spinner } from '@/src/components/ui/spinner';

export default function LoginPage() {
  const { login, loading } = useAuth();
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await login(username, password);
      router.push('/problems');
    } catch (err: any) {
      setError(err?.message || '登录失败');
    }
  }

  return (
    <div className="mx-auto max-w-sm py-20 w-full">
      <h1 className="text-2xl font-semibold mb-6">登录</h1>
      <form onSubmit={onSubmit} className="space-y-4">
        <div>
          <label className="block mb-1 text-sm">用户名</label>
          <Input value={username} onChange={e => setUsername(e.target.value)} required autoFocus />
        </div>
        <div>
          <label className="block mb-1 text-sm">密码</label>
          <Input type="password" value={password} onChange={e => setPassword(e.target.value)} required />
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <Button type="submit" disabled={loading} className="w-full">
          {loading && <Spinner className="mr-2" />} 登录
        </Button>
      </form>
    </div>
  );
}
