import { lazy, Suspense } from 'react'
import { createBrowserRouter } from 'react-router'
import { RoleRoute, ProtectedRoute } from './guards'
import { NotFoundPage } from '@/components/NotFoundPage'

const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage').then((module) => ({ default: module.LoginPage })))
const RegisterPage = lazy(() => import('@/features/auth/pages/RegisterPage').then((module) => ({ default: module.RegisterPage })))
const UserHomePage = lazy(() => import('@/features/user/pages/UserHomePage').then((module) => ({ default: module.UserHomePage })))
const TenantDashboardPage = lazy(() => import('@/features/tenant/pages/TenantDashboardPage').then((module) => ({ default: module.TenantDashboardPage })))
const AdminDashboardPage = lazy(() => import('@/features/admin/pages/AdminDashboardPage').then((module) => ({ default: module.AdminDashboardPage })))

const fallback = <div className="route-loader">Loading…</div>

export const router = createBrowserRouter([
  { path: '/login', element: <Suspense fallback={fallback}><LoginPage /></Suspense> },
  { path: '/register', element: <Suspense fallback={fallback}><RegisterPage /></Suspense> },
  {
    element: <ProtectedRoute />,
    children: [
      { path: '/', element: <Suspense fallback={fallback}><UserHomePage /></Suspense> },
      { path: '/user', element: <Suspense fallback={fallback}><UserHomePage /></Suspense> }
    ]
  },
  {
    element: <RoleRoute allowedRoles={['owner', 'staff']} />,
    children: [{ path: '/tenant', element: <Suspense fallback={fallback}><TenantDashboardPage /></Suspense> }]
  },
  {
    element: <RoleRoute allowedRoles={['admin']} />,
    children: [{ path: '/admin', element: <Suspense fallback={fallback}><AdminDashboardPage /></Suspense> }]
  },
  { path: '*', element: <NotFoundPage /> }
])
