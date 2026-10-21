import { lazy, Suspense } from 'react'
import { createBrowserRouter } from 'react-router'
import { ProtectedRoute } from './guards'
import { NotFoundPage } from '@/components/NotFoundPage'
import type { Portal } from '@/app/portal'

const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage').then((module) => ({ default: module.LoginPage })))
const RegisterPage = lazy(() => import('@/features/auth/pages/RegisterPage').then((module) => ({ default: module.RegisterPage })))
const UserHomePage = lazy(() => import('@/features/user/pages/UserHomePage').then((module) => ({ default: module.UserHomePage })))
const TenantDashboardPage = lazy(() => import('@/features/tenant/pages/TenantDashboardPage').then((module) => ({ default: module.TenantDashboardPage })))
const AdminDashboardPage = lazy(() => import('@/features/admin/pages/AdminDashboardPage').then((module) => ({ default: module.AdminDashboardPage })))

const fallback = <div className="route-loader">Loading…</div>

export function router(portal: Portal) {
  const protectedPage = portal === 'user'
    ? <Suspense fallback={fallback}><UserHomePage /></Suspense>
    : portal === 'tenant'
      ? <Suspense fallback={fallback}><TenantDashboardPage /></Suspense>
      : <Suspense fallback={fallback}><AdminDashboardPage /></Suspense>

  const routes = [
    { path: '/login', element: <Suspense fallback={fallback}><LoginPage /></Suspense> },
    ...(portal === 'user' ? [{ path: '/register', element: <Suspense fallback={fallback}><RegisterPage /></Suspense> }] : []),
    { element: <ProtectedRoute />, children: [{ path: '/', element: protectedPage }] },
    { path: '*', element: <NotFoundPage /> }
  ]

  return createBrowserRouter(routes)
}
