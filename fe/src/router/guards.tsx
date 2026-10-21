import { Navigate, Outlet, useLocation } from 'react-router'
import { useAuthStore } from '@/features/auth/auth.store'
import type { UserRole } from '@/features/auth/auth.types'
import { useI18n } from '@/lib/use-i18n'

export function ProtectedRoute() {
  const user = useAuthStore((state) => state.user)
  const location = useLocation()
  return user ? <Outlet /> : <Navigate to="/login" replace state={{ from: location.pathname }} />
}

function getUserRoles(roles: UserRole[] | undefined) {
  return new Set(roles ?? [])
}

export function RoleRoute({ allowedRoles }: { allowedRoles: UserRole[] }) {
  const user = useAuthStore((state) => state.user)
  const location = useLocation()

  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />

  const roles = getUserRoles(user.roles)
  return allowedRoles.some((role) => roles.has(role)) ? <Outlet /> : <AccessDeniedPage />
}

function AccessDeniedPage() {
  const { t } = useI18n()
  return (
    <main className="access-denied">
      <div className="access-denied-card">
        <p className="eyebrow">{t('accessDenied.eyebrow')}</p>
        <h1>{t('accessDenied.title')}</h1>
        <p>{t('accessDenied.description')}</p>
        <a className="primary-button" href="/">{t('auth.backToUser')}</a>
      </div>
    </main>
  )
}
