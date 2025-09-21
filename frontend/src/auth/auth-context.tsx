"use client";
import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { apiPost } from '@/src/api/client';

interface User {
  id: string;
  username: string;
  roles: string[];
}

interface AuthState {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  hasRole: (role: string) => boolean;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  // TODO: 可在此加入刷新 token 或 /me 请求
  useEffect(() => {
    const cached = localStorage.getItem('user');
    if (cached) {
      try { setUser(JSON.parse(cached)); } catch { /* ignore */ }
    }
    const handler = () => {
      logout();
      router.replace('/login');
    };
    window.addEventListener('auth:unauthorized', handler as EventListener);
    return () => window.removeEventListener('auth:unauthorized', handler as EventListener);
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    setLoading(true);
    try {
      const data = await apiPost<{ user: User; tokens: { access_token: string } }>("/auth/login", { username, password }, { auth: false });
      localStorage.setItem('access_token', data.tokens.access_token);
      localStorage.setItem('user', JSON.stringify(data.user));
      setUser(data.user);
    } finally {
      setLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    setUser(null);
  }, []);

  const hasRole = useCallback((role: string) => {
    return !!user?.roles?.includes(role);
  }, [user]);

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, hasRole }}>
      {children}
    </AuthContext.Provider>
  );
};

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
