"use client";
import React, {useState, useEffect, useRef} from 'react';
import {useParams, useRouter} from 'next/navigation';
import {useProblem} from '../../../../src/hooks/useProblem';
import {useCreateSubmission} from '../../../../src/hooks/useCreateSubmission';
import {Button} from '../../../../src/components/ui/button';
import {CodeEditor} from '../../../../src/components/code/CodeEditor';
import {ApiError} from '../../../../src/api/client';
import {useToast} from '../../../../src/components/ui/toast';
import {toUserMessage} from '../../../../src/lib/error';
import {getTemplate, isSupportedLanguage, SupportedLanguage} from '../../../../src/lib/codeTemplates';
import { useRequireAuth } from '@/src/hooks/useRequireAuth';

const MAX_CODE_LENGTH = 100_000; // 与后端限制保持一致（若需同步可提取配置）

export default function SubmitPage() {
  const params = useParams<{id: string}>();
  const router = useRouter();
  const {data: problem} = useProblem(params?.id);
  useRequireAuth();
  const {mutate, isPending, error} = useCreateSubmission();
  const {push: pushToast} = useToast();

  const [language, setLanguage] = useState<SupportedLanguage>('cpp');
  const [code, setCode] = useState('');
  const [tooLong, setTooLong] = useState(false);
  const [conflictMsg, setConflictMsg] = useState<string | null>(null);
  const langChangedRef = useRef(false);

  // 初始化：从 localStorage 读取上次使用语言
  useEffect(() => {
    const stored = typeof window !== 'undefined' ? localStorage.getItem('submit_language') : null;
    if (stored && isSupportedLanguage(stored)) {
      setLanguage(stored);
    }
  }, []);

  // 语言变更时持久化
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('submit_language', language);
    }
  }, [language]);

  function applyTemplate() {
    const tpl = getTemplate(language);
    setCode(tpl);
    pushToast({ variant: 'success', title: '已填充模板', description: language.toUpperCase() + ' 模板已插入' });
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!params?.id) return;
    if (code.length > MAX_CODE_LENGTH) {
      setTooLong(true);
      return;
    }
    setTooLong(false);
    setConflictMsg(null);
    mutate({problemId: params.id, language, code}, {
      onSuccess: () => {
        pushToast({ variant: 'success', title: '提交成功', description: '已创建提交，稍后可查看判题状态' });
        router.push('/submissions');
      },
      onError: (err: ApiError) => {
        if (err.conflict) {
          setConflictMsg('提交冲突，请刷新或稍后重试。');
          pushToast({ variant: 'warning', title: '提交冲突', description: '请刷新页面后重试' });
          return;
        }
        const msg = toUserMessage(err);
        pushToast({ variant: 'error', title: msg.title, description: msg.description });
      }
    });
  }

  return (
    <div className="p-6 max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">提交代码 - {problem?.title || params?.id}</h1>
        <Button variant="ghost" size="sm" onClick={() => router.back()}>返回</Button>
      </div>
      <form onSubmit={onSubmit} className="space-y-4">
        <div className="space-y-1">
          <label className="text-sm font-medium">语言</label>
          <select
            className="border rounded px-2 py-1 text-sm bg-white"
            value={language}
            onChange={e => {
              const next = e.target.value;
              if (isSupportedLanguage(next)) {
                setLanguage(next);
                langChangedRef.current = true;
              }
            }}
          >
            <option value="cpp">C++</option>
            <option value="python">Python</option>
            <option value="go">Go</option>
            <option value="java">Java</option>
          </select>
          <div className="pt-2 flex gap-2">
            <Button type="button" variant="outline" size="sm" onClick={applyTemplate}>填充模板</Button>
            {langChangedRef.current && code.trim().length === 0 && (
              <span className="text-xs text-neutral-500 self-center">语言已切换，可点击“填充模板”快速插入</span>
            )}
          </div>
        </div>
        <div className="space-y-1">
          <label className="text-sm font-medium flex items-center justify-between">
            <span>代码</span>
            <span className={`text-xs ${code.length > MAX_CODE_LENGTH * 0.9 ? 'text-orange-600' : 'text-gray-400'}`}>{code.length}/{MAX_CODE_LENGTH}</span>
          </label>
          <CodeEditor
            language={language}
            value={code}
            onChange={v => setCode(v)}
            height={400}
          />
          {tooLong && <p className="text-xs text-red-600">代码长度超过允许的最大值 {MAX_CODE_LENGTH} 字符。</p>}
        </div>
        {error && !error.conflict && !tooLong && (
          <div className="text-sm text-red-600">提交失败: {toUserMessage(error).title}</div>
        )}
        {conflictMsg && (
          <div className="text-sm text-amber-600 border border-amber-300 bg-amber-50 px-3 py-2 rounded">{conflictMsg}</div>
        )}
        <div className="flex gap-2">
          <Button type="submit" disabled={isPending || tooLong}>{isPending ? '提交中...' : '提交'}</Button>
          <Button type="button" variant="outline" onClick={() => router.back()}>取消</Button>
        </div>
      </form>
    </div>
  );
}
