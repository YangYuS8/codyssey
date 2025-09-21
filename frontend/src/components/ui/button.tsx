"use client";
import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/src/lib/utils';

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-60 disabled:pointer-events-none ring-offset-background',
  {
    variants: {
      variant: {
        default: 'bg-black text-white hover:bg-black/80 dark:bg-white dark:text-black dark:hover:bg-white/80',
        outline: 'border border-neutral-300 dark:border-neutral-600 hover:bg-neutral-100 dark:hover:bg-neutral-800',
        ghost: 'hover:bg-neutral-100 dark:hover:bg-neutral-800',
      },
      size: {
        sm: 'h-8 px-3 rounded-md',
        md: 'h-9 px-4 rounded-md',
        lg: 'h-10 px-6 rounded-md',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
    },
  }
);

type ButtonVariantProps = VariantProps<typeof buttonVariants>;

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariantProps['variant'];
  size?: ButtonVariantProps['size'];
  asChild?: boolean; // 预留未来与 Radix Slot 集成
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild, ...props }, ref) => {
    // 目前不实现 asChild，仅保留 API 兼容位
    return <button ref={ref} className={cn(buttonVariants({ variant, size }), className)} {...props} />;
  }
);
Button.displayName = 'Button';
