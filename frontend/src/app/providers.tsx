"use client";
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { AuthProvider } from '@/src/auth/auth-context';
import { Toaster } from '@/src/components/ui/toaster';

const qc = new QueryClient({
  defaultOptions: {
    queries: {
      retry(failureCount, error: unknown) {
        const httpStatus = (error as { httpStatus?: number })?.httpStatus;
        if (httpStatus && httpStatus < 500) return false;
        return failureCount < 2;
      },
      refetchOnWindowFocus: false,
    },
  },
});

export const AppProviders: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <QueryClientProvider client={qc}>
      <AuthProvider>
        {children}
        <Toaster />
      </AuthProvider>
    </QueryClientProvider>
  );
};
