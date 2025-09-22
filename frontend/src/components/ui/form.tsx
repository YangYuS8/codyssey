"use client";
import * as React from 'react';
import { FieldError, UseFormRegisterReturn } from 'react-hook-form';
import { cn } from '@/src/lib/utils';

interface FormFieldProps extends React.HTMLAttributes<HTMLDivElement> {
  label?: string;
  error?: FieldError | string;
  requiredMark?: boolean;
  children: React.ReactNode;
}

export const FormField: React.FC<FormFieldProps> = ({ label, error, requiredMark, className, children }) => {
  return (
    <div className={cn('space-y-1', className)}>
      {label && (
        <label className="block text-sm font-medium">
          {label}{requiredMark && <span className="text-red-600 ml-0.5">*</span>}
        </label>
      )}
      {children}
      {error && (
        <p className="text-xs text-red-600">{typeof error === 'string' ? error : error.message}</p>
      )}
    </div>
  );
};

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  register?: UseFormRegisterReturn;
}
export const FormInput: React.FC<InputProps> = ({ register, className, ...rest }) => {
  return <input {...register} {...rest} className={cn('border rounded px-2 py-1 text-sm w-full', className)} />;
};
