"use client";
import React from 'react';
import { cn } from '@/src/lib/utils';

export interface StatusBadgeProps {
  status: string | undefined | null;
  className?: string;
  size?: 'sm' | 'md';
}

const colorMap: Record<string, string> = {
  pending: 'bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-200',
  running: 'bg-blue-100 text-blue-700 dark:bg-blue-700/40 dark:text-blue-200',
  accepted: 'bg-green-100 text-green-700 dark:bg-green-700/40 dark:text-green-200',
  rejected: 'bg-red-100 text-red-700 dark:bg-red-700/40 dark:text-red-200',
  failed: 'bg-orange-100 text-orange-700 dark:bg-orange-700/40 dark:text-orange-200',
};

function formatStatus(s?: string | null) {
  if (!s) return '-';
  return s.charAt(0).toUpperCase() + s.slice(1);
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status, className, size = 'sm' }) => {
  const s = (status || '').toLowerCase();
  const color = colorMap[s] || 'bg-neutral-200 text-neutral-600 dark:bg-neutral-700 dark:text-neutral-300';
  const base = size === 'sm' ? 'text-[11px] px-2 py-0.5' : 'text-xs px-2.5 py-1';
  return (
    <span className={cn('inline-flex items-center rounded-full font-medium tracking-wide', base, color, className)}>
      {formatStatus(s)}
    </span>
  );
};
