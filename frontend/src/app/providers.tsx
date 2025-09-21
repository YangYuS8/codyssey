"use client";
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { AuthProvider } from '@/src/auth/auth-context';

const qc = new QueryClient({
  defaultOptions: {
    queries: {
      retry(failureCount, error: any) {
        if (error?.httpStatus && error.httpStatus < 500) return false;
        return failureCount < 2;
      },
      refetchOnWindowFocus: false,
    },
  },
});

export const AppProviders: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <QueryClientProvider client={qc}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  );
};
