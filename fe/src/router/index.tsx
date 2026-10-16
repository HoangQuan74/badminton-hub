import { lazy, Suspense } from 'react'
import { createBrowserRouter } from 'react-router'
import { ProtectedRoute } from './guards'

const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage').then((module) => ({ default: module.LoginPage })))
const RegisterPage = lazy(() => import('@/features/auth/pages/RegisterPage').then((module) => ({ default: module.RegisterPage })))
const HomePage = lazy(() => import('@/features/home/HomePage').then((module) => ({ default: module.HomePage })))

const fallback = <div className="route-loader">Đang tải…</div>

export const router = createBrowserRouter([
  { path: '/login', element: <Suspense fallback={fallback}><LoginPage /></Suspense> },
  { path: '/register', element: <Suspense fallback={fallback}><RegisterPage /></Suspense> },
  {
    element: <ProtectedRoute />,
    children: [{ path: '/', element: <Suspense fallback={fallback}><HomePage /></Suspense> }]
  },
  { path: '*', element: <div className="not-found"><h1>404</h1><p>Trang bạn tìm không tồn tại.</p><a href="/">Về trang chủ</a></div> }
])
