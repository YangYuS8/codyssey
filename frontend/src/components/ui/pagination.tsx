"use client";
import React from 'react';
import { cn } from '@/src/lib/utils';

interface PaginationProps {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (p: number) => void;
  className?: string;
  disabled?: boolean;
}

export const Pagination: React.FC<PaginationProps> = ({ page, pageSize, total, onPageChange, className, disabled }) => {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const canPrev = page > 1;
  const canNext = page < totalPages;
  return (
    <div className={cn('flex items-center gap-2 text-xs text-neutral-600', className)}>
      <span>第 {page} / {totalPages} 页 · 共 {total} 条</span>
      <button
        type="button"
        disabled={!canPrev || disabled}
        onClick={() => canPrev && onPageChange(page - 1)}
        className="px-2 py-1 border rounded disabled:opacity-40"
      >上一页</button>
      <button
        type="button"
        disabled={!canNext || disabled}
        onClick={() => canNext && onPageChange(page + 1)}
        className="px-2 py-1 border rounded disabled:opacity-40"
      >下一页</button>
    </div>
  );
};
