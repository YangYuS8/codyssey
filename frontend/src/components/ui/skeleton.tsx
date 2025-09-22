"use client";
import React from 'react';
import { cn } from '@/src/lib/utils';

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  shimmer?: boolean;
}

export const Skeleton: React.FC<SkeletonProps> = ({ className, shimmer = true, ...rest }) => (
  <div
    className={cn(
      'bg-neutral-200 dark:bg-neutral-700 rounded-md relative overflow-hidden',
      shimmer && 'after:absolute after:inset-0 after:-translate-x-full after:animate-[shimmer_1.5s_infinite] after:bg-gradient-to-r after:from-transparent after:via-white/60 dark:after:via-white/10 after:to-transparent',
      className
    )}
    {...rest}
  />
);

// Tailwind 未必有此自定义动画，可在全局 CSS 中追加（若尚未定义）:
// @keyframes shimmer { 100% { transform: translateX(100%); } }
