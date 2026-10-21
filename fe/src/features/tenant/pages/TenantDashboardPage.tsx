import { Building2, CalendarDays, ChevronRight, CircleHelp, LayoutDashboard, Settings2, UsersRound } from 'lucide-react'
import { Link } from 'react-router'
import { tokenStorage } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'
import { authApi } from '@/features/auth/auth.api'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { queryClient } from '@/lib/query-client'
import { PortalShell } from '@/features/portals/components/PortalShell'
import { useI18n } from '@/lib/use-i18n'

export function TenantDashboardPage() {
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
    { label: t('portal.overview'), href: '/tenant', icon: LayoutDashboard },
    { label: t('portal.venuesAndCourts'), href: '/tenant#venues', icon: Building2 },
    { label: t('portal.schedule'), href: '/tenant#bookings', icon: CalendarDays },
    { label: t('portal.staff'), href: '/tenant#staff', icon: UsersRound }
  ]

  return (
    <PortalShell
      portal="Tenant"
      title={t('portal.tenantTitle')}
      navigation={navigation}
      userName={user.full_name}
      onLogout={() => logout.mutate()}
      logoutPending={logout.isPending}
    >
      <section className="portal-intro">
        <div>
          <p className="eyebrow">{t('portal.tenantEyebrow')}</p>
          <h1>{t('portal.tenantHeading')}</h1>
          <p>{t('portal.tenantDescription')}</p>
        </div>
        <button className="primary-button" type="button" disabled><Building2 size={18} /> {t('portal.createVenue')}</button>
      </section>

      <div className="portal-notice" role="status">
        <CircleHelp size={18} /> {t('portal.tenantNotice')}
      </div>

      <section className="metric-grid" aria-label="Tổng quan tenant">
        <article className="metric-card"><span className="metric-icon"><Building2 size={18} /></span><p>{t('portal.venues')}</p><strong>—</strong><small>{t('portal.noData')}</small></article>
        <article className="metric-card"><span className="metric-icon"><CalendarDays size={18} /></span><p>{t('portal.todayBookings')}</p><strong>—</strong><small>{t('portal.noData')}</small></article>
        <article className="metric-card"><span className="metric-icon"><UsersRound size={18} /></span><p>{t('portal.staff')}</p><strong>—</strong><small>{t('portal.noData')}</small></article>
      </section>

      <section className="portal-panel" id="venues">
        <div className="panel-heading"><div><p className="eyebrow">{t('portal.organizationScope')}</p><h2>{t('portal.venuesAndCourts')}</h2></div><button className="text-button" type="button" disabled>{t('portal.viewList')} <ChevronRight size={17} /></button></div>
        <div className="portal-empty"><Building2 size={24} /><h3>{t('portal.noVenues')}</h3><p>{t('portal.noVenuesDescription')}</p></div>
      </section>

      <section className="portal-panel two-column-panels" id="bookings">
        <div><p className="eyebrow">{t('portal.operations')}</p><h2>{t('portal.schedule')}</h2><p className="panel-muted">{t('portal.scheduleDescription')}</p><Link className="panel-link" to="/tenant#bookings">{t('portal.openSchedule')} <ChevronRight size={16} /></Link></div>
        <div id="staff"><p className="eyebrow">{t('portal.access')}</p><h2>{t('portal.staff')}</h2><p className="panel-muted">{t('portal.staffDescription')}</p><button className="panel-link" type="button" disabled>{t('portal.inviteStaff')} <ChevronRight size={16} /></button></div>
      </section>

      <button className="portal-settings" type="button" disabled><Settings2 size={17} /> {t('portal.tenantSettings')}</button>
    </PortalShell>
  )
}
