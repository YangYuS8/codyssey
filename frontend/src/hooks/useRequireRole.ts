"use client";
import { useAuth } from '@/src/auth/auth-context';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export function useRequireRole(role: string, redirect = '/login') {
  const { user } = useAuth();
  const router = useRouter();
  useEffect(() => {
    if (!user) return; // 让 useRequireAuth 先处理未登录
    if (role && !user.roles.includes(role)) {
      router.replace(redirect + '?denied=1');
    }
  }, [user, role, redirect, router]);
}
