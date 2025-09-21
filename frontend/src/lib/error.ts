import { ApiError } from '../api/client';

// 用户可读消息映射层
// 约定：ApiError.code 为后端统一错误码；如果后端演进只需在此集中修改

export function toUserMessage(err: unknown): { title: string; description?: string } {
  if (!err) return { title: '未知错误' };
  if ((err as any).issues) {
    return { title: '数据格式错误', description: '返回数据结构与前端期望不一致' };
  }
  const e = err as ApiError;
  if (!e.code) return { title: '请求失败', description: (e as any).message || '未知错误' };
  switch (e.code) {
    case 'UNAUTHORIZED':
      return { title: '未登录', description: '请重新登录' };
    case 'FORBIDDEN':
      return { title: '无权限', description: '当前账号无操作权限' };
    case 'CONFLICT':
      return { title: '资源冲突', description: '请刷新后重试' };
    case 'PAYLOAD_TOO_LARGE':
      return { title: '内容过大', description: '请求体或代码长度超出限制' };
    case 'NOT_FOUND':
    case 'PROBLEM_NOT_FOUND':
    case 'SUBMISSION_NOT_FOUND':
      return { title: '未找到资源', description: '资源可能已被删除' };
    default:
      if (e.httpStatus >= 500) return { title: '服务器异常', description: '请稍后再试' };
      return { title: '请求失败', description: e.message };
  }
}
