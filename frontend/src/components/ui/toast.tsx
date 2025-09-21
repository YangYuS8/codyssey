"use client";
import React, { createContext, useContext, useCallback, useState } from 'react';
import { cn } from '@/src/lib/utils';

export interface ToastOptions {
  id?: string;
  title?: string;
  description?: string;
  variant?: 'default' | 'success' | 'error' | 'warning';
  duration?: number; // ms
}

interface ToastInternal extends ToastOptions {
  id: string;
  createdAt: number;
}

interface ToastContextValue {
  toasts: ToastInternal[];
  push: (opts: ToastOptions) => void;
  dismiss: (id: string) => void;
  clear: () => void;
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastInternal[]>([]);

  const dismiss = useCallback((id: string) => {
    setToasts(ts => ts.filter(t => t.id !== id));
  }, []);

  const clear = useCallback(() => setToasts([]), []);

  const push = useCallback((opts: ToastOptions) => {
    const id = opts.id || Math.random().toString(36).slice(2);
    const t: ToastInternal = {
      id,
      title: opts.title,
      description: opts.description,
      variant: opts.variant || 'default',
      duration: opts.duration ?? 4000,
      createdAt: Date.now(),
    };
    setToasts(ts => [...ts, t]);
    if (t.duration && t.duration > 0) {
      setTimeout(() => {
        setToasts(ts => ts.filter(tt => tt.id !== id));
      }, t.duration);
    }
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, push, dismiss, clear }}>
      {children}
    </ToastContext.Provider>
  );
};

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within ToastProvider');
  return ctx;
}

export const ToastViewport: React.FC = () => {
  const { toasts, dismiss } = useToast();
  return (
    <div className="fixed bottom-4 right-4 flex flex-col gap-2 z-50 max-w-sm">
      {toasts.map(t => (
        <div
          key={t.id}
          className={cn(
            'rounded border p-3 shadow bg-white dark:bg-neutral-900 text-sm animate-in fade-in slide-in-from-bottom-2',
            t.variant === 'success' && 'border-green-300',
            t.variant === 'error' && 'border-red-300',
            t.variant === 'warning' && 'border-amber-300'
          )}
        >
          <div className="flex justify-between gap-4">
            <div className="space-y-1">
              {t.title && <div className="font-medium">{t.title}</div>}
              {t.description && <div className="text-xs text-neutral-600 dark:text-neutral-300 leading-snug">{t.description}</div>}
            </div>
            <button
              onClick={() => dismiss(t.id)}
              className="text-xs text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200"
              aria-label="Close"
            >×</button>
          </div>
        </div>
      ))}
    </div>
  );
};
