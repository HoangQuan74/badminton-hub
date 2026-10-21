import { AlertTriangle, Building2, CheckCircle2, ClipboardCheck, LayoutDashboard, ShieldCheck, UsersRound } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { authApi } from '@/features/auth/auth.api'
import { tokenStorage } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'
import { queryClient } from '@/lib/query-client'
import { PortalShell } from '@/features/portals/components/PortalShell'
import { useI18n } from '@/lib/use-i18n'

export function AdminDashboardPage() {
  const { t } = useI18n()
  const user = useAuthStore((state) => state.user)!
  const reset = useAuthStore((state) => state.reset)
  const navigate = useNavigate()
  const logout = useMutation({
    mutationFn: authApi.logout,
    onSettled: () => {
      tokenStorage.clear()
      queryClient.clear()
      reset()
      navigate('/login', { replace: true })
    }
  })
  const navigation = [
    { label: t('portal.overview'), href: '/admin', icon: LayoutDashboard },
    { label: t('portal.reviewQueue'), href: '/admin#organizations', icon: ClipboardCheck },
    { label: t('portal.venuesAndCourts'), href: '/admin#venues', icon: Building2 },
    { label: t('portal.accessControl'), href: '/admin#users', icon: UsersRound }
  ]

  return (
    <PortalShell
      portal="Admin"
      title={t('portal.adminTitle')}
      navigation={navigation}
      userName={user.full_name}
      onLogout={() => logout.mutate()}
      logoutPending={logout.isPending}
    >
      <section className="portal-intro">
        <div>
          <p className="eyebrow">{t('portal.adminEyebrow')}</p>
          <h1>{t('portal.adminHeading')}</h1>
          <p>{t('portal.adminDescription')}</p>
        </div>
        <button className="secondary-button" type="button" disabled><ShieldCheck size={18} /> {t('portal.platformRules')}</button>
      </section>

      <div className="portal-notice warning" role="status">
        <AlertTriangle size={18} /> {t('portal.adminNotice')}
      </div>

      <section className="metric-grid" aria-label="Tổng quan admin">
        <article className="metric-card"><span className="metric-icon"><ClipboardCheck size={18} /></span><p>{t('portal.pendingApproval')}</p><strong>—</strong><small>Organization and venue</small></article>
        <article className="metric-card"><span className="metric-icon"><Building2 size={18} /></span><p>{t('portal.activeVenues')}</p><strong>—</strong><small>{t('portal.noData')}</small></article>
        <article className="metric-card"><span className="metric-icon"><CheckCircle2 size={18} /></span><p>{t('portal.paymentsToReview')}</p><strong>—</strong><small>{t('portal.noData')}</small></article>
      </section>

      <section className="portal-panel" id="organizations">
        <div className="panel-heading"><div><p className="eyebrow">{t('portal.governance')}</p><h2>{t('portal.reviewQueue')}</h2></div><button className="text-button" type="button" disabled>{t('portal.openQueue')} <span aria-hidden="true">→</span></button></div>
        <div className="portal-empty"><ClipboardCheck size={24} /><h3>{t('portal.noApprovalData')}</h3><p>{t('portal.noApprovalDescription')}</p></div>
      </section>

      <section className="portal-panel two-column-panels" id="venues">
        <div><p className="eyebrow">{t('portal.monitoring')}</p><h2>{t('portal.venuesAndCourts')}</h2><p className="panel-muted">{t('portal.venueBookingDescription')}</p><button className="panel-link" type="button" disabled>{t('portal.openMonitoring')} <span aria-hidden="true">→</span></button></div>
        <div id="users"><p className="eyebrow">{t('portal.accessControl')}</p><h2>{t('portal.accessControl')}</h2><p className="panel-muted">{t('portal.userAccessDescription')}</p><button className="panel-link" type="button" disabled>{t('portal.openAccessReview')} <span aria-hidden="true">→</span></button></div>
      </section>

      <button className="portal-settings" type="button" disabled><ShieldCheck size={17} /> {t('portal.platformSettings')}</button>
    </PortalShell>
  )
}
