"use client";
import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';

// 简单客户端守卫：若无 access_token 则跳转 /login 并附带 next 参数
export function useRequireAuth(enabled: boolean = true) {
  const router = useRouter();
  const pathname = usePathname();
  useEffect(() => {
    if (!enabled) return;
    const token = typeof window !== 'undefined' ? localStorage.getItem('access_token') : null;
    if (!token) {
      const next = encodeURIComponent(pathname || '/');
      router.replace(`/login?next=${next}`);
    }
  }, [router, pathname, enabled]);
}
