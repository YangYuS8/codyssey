import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// 受保护路由前缀（可根据需要扩展）
const PROTECTED_PREFIXES = ['/problems', '/submissions'];
const LOGIN_PATH = '/login';

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const token = req.cookies.get('auth_token')?.value;

  const isProtected = PROTECTED_PREFIXES.some(p => pathname.startsWith(p));
  const isLogin = pathname === LOGIN_PATH;

  if (isProtected && !token) {
    const url = req.nextUrl.clone();
    url.pathname = LOGIN_PATH;
    url.searchParams.set('next', pathname);
    return NextResponse.redirect(url);
  }

  if (isLogin && token) {
    const url = req.nextUrl.clone();
    url.pathname = '/problems';
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/login',
    '/problems/:path*',
    '/submissions/:path*'
  ]
};