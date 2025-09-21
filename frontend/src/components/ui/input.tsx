"use client";
import * as React from 'react';
import { cn } from '@/src/lib/utils';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn('flex h-9 w-full rounded-md border border-neutral-300 dark:border-neutral-700 bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-black dark:focus-visible:ring-white disabled:opacity-50 disabled:cursor-not-allowed', className)}
      {...props}
    />
  )
);
Input.displayName = 'Input';
