// 基础 API 客户端封装，适配后端统一 envelope { data, error }

export interface ApiError extends Error {
  code: string;
  httpStatus: number;
  conflict?: boolean;
  unauthorized?: boolean;
  forbidden?: boolean;
  payloadTooLarge?: boolean;
  notFound?: boolean;
}

interface RequestOptions<B = unknown> {
  method?: string;
  body?: B;
  auth?: boolean; // 默认 true
  headers?: Record<string, string>;
  signal?: AbortSignal;
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || '';

function buildError(code: string, status: number, message: string): ApiError {
  const e = new Error(message) as ApiError;
  e.code = code;
  e.httpStatus = status;
  e.conflict = code === 'CONFLICT';
  e.unauthorized = code === 'UNAUTHORIZED';
  e.forbidden = code === 'FORBIDDEN';
  e.payloadTooLarge = code === 'PAYLOAD_TOO_LARGE';
  e.notFound = code.endsWith('_NOT_FOUND') || code === 'NOT_FOUND';
  return e;
}

export async function apiFetch<T, B = unknown>(path: string, opts: RequestOptions<B> = {}): Promise<T> {
  const { method = 'GET', body, auth = true, headers = {}, signal } = opts;
  const finalHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    ...headers,
  };

  // token 简单示例：从 localStorage 读取
  if (auth) {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token');
      if (token) finalHeaders['Authorization'] = `Bearer ${token}`;
    }
  }

  const res = await fetch(API_BASE + path, {
    method,
    body: body ? JSON.stringify(body) : undefined,
    headers: finalHeaders,
    signal,
    credentials: 'include',
  });

  const contentType = res.headers.get('content-type');
  let json: unknown = null;
  if (contentType && contentType.includes('application/json')) {
    json = await res.json().catch(() => null);
  } else if (!res.ok) {
    throw buildError('UNKNOWN', res.status, `HTTP ${res.status}`);
  }
  const envelope = json as { data?: T; error?: { code?: string; message?: string } } | null;
  if (res.ok) {
    return (envelope && envelope.data !== undefined ? envelope.data : (json as T)) as T;
  }
  const code = envelope?.error?.code || 'UNKNOWN';
  const message = envelope?.error?.message || `HTTP ${res.status}`;
  const err = buildError(code, res.status, message);
  if (err.unauthorized && typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('auth:unauthorized'));
  }
  throw err;
}

// 常用 GET 简化
export function apiGet<T>(path: string, opts: Omit<RequestOptions<never>, 'method' | 'body'> = {}) {
  return apiFetch<T>(path, { ...opts, method: 'GET' });
}

export function apiPost<T, B = unknown>(path: string, body?: B, opts: Omit<RequestOptions<B>, 'method' | 'body'> = {}) {
  return apiFetch<T, B>(path, { ...opts, method: 'POST', body });
}

export function apiPatch<T, B = unknown>(path: string, body?: B, opts: Omit<RequestOptions<B>, 'method' | 'body'> = {}) {
  return apiFetch<T, B>(path, { ...opts, method: 'PATCH', body });
}
