"use client";
import React from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/src/auth/auth-context';
import { Button } from '@/src/components/ui/button';
import { Spinner } from '@/src/components/ui/spinner';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { FormField, FormInput } from '@/src/components/ui/form';

const LoginSchema = z.object({
  username: z.string().min(1, '请输入用户名'),
  password: z.string().min(1, '请输入密码'),
});
type LoginForm = z.infer<typeof LoginSchema>;

export default function LoginPage() {
  const { login, loading } = useAuth();
  const router = useRouter();
  const { register, handleSubmit, formState: { errors }, setError } = useForm<LoginForm>({
    resolver: zodResolver(LoginSchema),
    mode: 'onBlur'
  });

  const onSubmit = async (values: LoginForm) => {
    try {
      await login(values.username, values.password);
      router.push('/problems');
    } catch (err: unknown) {
      const message = (err as { message?: string })?.message || '登录失败';
      setError('username', { message });
    }
  };

  return (
    <div className="mx-auto max-w-sm py-20 w-full">
      <h1 className="text-2xl font-semibold mb-6">登录</h1>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <FormField label="用户名" requiredMark error={errors.username}>
          <FormInput autoFocus placeholder="用户名" register={register('username')} />
        </FormField>
        <FormField label="密码" requiredMark error={errors.password}>
          <FormInput type="password" placeholder="密码" register={register('password')} />
        </FormField>
        <Button type="submit" disabled={loading} className="w-full">
          {loading && <Spinner className="mr-2" />} 登录
        </Button>
      </form>
    </div>
  );
}
