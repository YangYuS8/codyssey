"use client";
import React from 'react';
import dynamic from 'next/dynamic';
import { Spinner } from '../ui/spinner';

// 动态导入 Monaco，避免 SSR 报错
type MonacoComponent = React.ComponentType<{
  value: string;
  language: string;
  theme?: string;
  options?: Record<string, unknown>;
  onChange?: (value: string | undefined) => void;
  height?: number | string;
}>;

const Monaco = dynamic(async () => {
  const mod = await import('@monaco-editor/react');
  return mod.default as MonacoComponent;
}, {
  ssr: false,
  loading: () => <div className="h-80 flex items-center justify-center border rounded"><Spinner /></div>
});

export interface CodeEditorProps {
  value: string;
  language: string;
  height?: number | string;
  onChange?: (v: string) => void;
  readOnly?: boolean;
}

export const CodeEditor: React.FC<CodeEditorProps> = ({ value, language, height = 400, onChange, readOnly }) => {
  // @monaco-editor/react 的 onChange 第一个参数就是 value
  return (
    <div className="border rounded overflow-hidden">
      <Monaco
        value={value}
        language={mapLanguage(language)}
  onChange={(v: string | undefined) => onChange?.(v || '')}
        theme="vs-dark"
        options={{
          readOnly,
          fontSize: 14,
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          automaticLayout: true,
          wordWrap: 'on',
        }}
        height={height}
      />
    </div>
  );
};

function mapLanguage(lang: string): string {
  switch (lang) {
    case 'cpp': return 'cpp';
    case 'python': return 'python';
    case 'go': return 'go';
    case 'java': return 'java';
    default: return 'plaintext';
  }
}
