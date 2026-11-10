import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Building2, CalendarDays, ChevronRight, CircleHelp, LayoutDashboard, Pencil, Plus, Settings2, Trash2, UsersRound } from 'lucide-react'
import { Link } from 'react-router'
import { tokenStorage } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'
import { authApi } from '@/features/auth/auth.api'
import { useNavigate } from 'react-router'
import { PortalShell } from '@/features/portals/components/PortalShell'
import { useI18n } from '@/lib/use-i18n'
import { catalogApi } from '@/features/catalog/catalog.api'

export function TenantDashboardPage() {
  const { t } = useI18n()
  const user = useAuthStore((state) => state.user)!
  const reset = useAuthStore((state) => state.reset)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const canManageCourts = user.roles?.includes('tenant_owner') ?? false
  const venuesQuery = useQuery({ queryKey: ['tenant-venues'], queryFn: catalogApi.listTenantVenues })
  const [selectedVenueId, setSelectedVenueId] = useState('')
  const [courtName, setCourtName] = useState('')
  const [slotInterval, setSlotInterval] = useState(30)
  const [editingCourtId, setEditingCourtId] = useState<string | null>(null)
  const selectedVenue = venuesQuery.data?.find((venue) => venue.id === selectedVenueId)
  const courtsQuery = useQuery({
    queryKey: ['tenant-courts', selectedVenueId],
    queryFn: () => catalogApi.listTenantCourts(selectedVenueId),
    enabled: Boolean(selectedVenueId)
  })
  useEffect(() => {
    if (!selectedVenueId && venuesQuery.data?.[0]) setSelectedVenueId(venuesQuery.data[0].id)
  }, [selectedVenueId, venuesQuery.data])
  const saveCourt = useMutation({
    mutationFn: async () => {
      const payload = { name: courtName, slot_interval_minutes: slotInterval, ...(editingCourtId ? { status: 'ACTIVE' as const } : {}) }
      return editingCourtId ? catalogApi.updateCourt(editingCourtId, payload) : catalogApi.createCourt(selectedVenueId, payload)
    },
    onSuccess: () => {
      setCourtName('')
      setSlotInterval(30)
      setEditingCourtId(null)
      void queryClient.invalidateQueries({ queryKey: ['tenant-venues'] })
      void queryClient.invalidateQueries({ queryKey: ['tenant-courts', selectedVenueId] })
    }
  })
  const archiveCourt = useMutation({
    mutationFn: catalogApi.archiveCourt,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['tenant-venues'] })
      void queryClient.invalidateQueries({ queryKey: ['tenant-courts', selectedVenueId] })
    }
  })
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
    { label: t('portal.overview'), href: '/', icon: LayoutDashboard },
    { label: t('portal.venuesAndCourts'), href: '/#venues', icon: Building2 },
    { label: t('portal.schedule'), href: '/#bookings', icon: CalendarDays },
    { label: t('portal.staff'), href: '/#staff', icon: UsersRound }
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

      <div className="portal-notice" role="status"><CircleHelp size={18} /> {venuesQuery.isError ? t('portal.tenantLoadError') : t('portal.tenantNotice')}</div>

      <section className="metric-grid" aria-label="Tổng quan tenant">
        <article className="metric-card">
          <span className="metric-icon"><Building2 size={18} /></span>
          <p>{t('portal.venues')}</p>
          <strong>{venuesQuery.data?.length ?? '—'}</strong>
          <small>{venuesQuery.isLoading ? t('portal.loading') : t('portal.connected')}</small>
        </article>
        <article className="metric-card">
          <span className="metric-icon"><CalendarDays size={18} /></span>
          <p>{t('portal.todayBookings')}</p>
          <strong>—</strong>
          <small>{t('portal.noData')}</small>
        </article>
        <article className="metric-card">
          <span className="metric-icon"><UsersRound size={18} /></span>
          <p>{t('portal.staff')}</p>
          <strong>—</strong>
          <small>{t('portal.noData')}</small>
        </article>
      </section>

      <section className="portal-panel" id="venues">
        <div className="panel-heading"><div><p className="eyebrow">{t('portal.organizationScope')}</p><h2>{t('portal.venuesAndCourts')}</h2></div></div>
        {venuesQuery.isLoading && <div className="portal-empty" role="status"><p>{t('portal.loading')}</p></div>}
        {!venuesQuery.isLoading && venuesQuery.data?.length === 0 && (
          <div className="portal-empty">
            <Building2 size={24} />
            <h3>{t('portal.noVenues')}</h3>
            <p>{t('portal.noVenuesDescription')}</p>
          </div>
        )}
        {venuesQuery.data && venuesQuery.data.length > 0 && <div className="catalog-manager">
          <label className="field-label" htmlFor="tenant-venue">{t('portal.selectVenue')}</label>
          <select id="tenant-venue" value={selectedVenueId} onChange={(event) => { setSelectedVenueId(event.target.value); setEditingCourtId(null); setCourtName('') }}>
            {venuesQuery.data.map((venue) => <option key={venue.id} value={venue.id}>{venue.name} · {venue.status}</option>)}
          </select>
          {selectedVenue && <p className="panel-muted">{selectedVenue.address}</p>}
          {canManageCourts && <form className="court-form" onSubmit={(event) => { event.preventDefault(); saveCourt.mutate() }}>
            <label className="field-label" htmlFor="court-name">{editingCourtId ? t('portal.editCourt') : t('portal.addCourt')}</label>
            <div className="court-form-row">
              <input id="court-name" value={courtName} onChange={(event) => setCourtName(event.target.value)} placeholder={t('portal.courtNamePlaceholder')} required maxLength={80} />
              <select value={slotInterval} onChange={(event) => setSlotInterval(Number(event.target.value))}>
                <option value={30}>30 {t('portal.minutes')}</option>
                <option value={60}>60 {t('portal.minutes')}</option>
                <option value={90}>90 {t('portal.minutes')}</option>
              </select>
              <button className="primary-button" type="submit" disabled={saveCourt.isPending}>
                {editingCourtId ? <Pencil size={16} /> : <Plus size={16} />}
                {saveCourt.isPending ? t('portal.saving') : t('portal.saveCourt')}
              </button>
              {editingCourtId && <button className="text-button" type="button" onClick={() => { setEditingCourtId(null); setCourtName('') }}>{t('portal.cancel')}</button>}
            </div>
          </form>}
          {saveCourt.isError && <p className="form-error" role="alert">{t('portal.courtSaveError')}</p>}
          {courtsQuery.isLoading && <p className="panel-muted">{t('portal.loading')}</p>}
          {courtsQuery.isError && <div className="portal-empty"><p>{t('portal.courtLoadError')}</p><button className="secondary-button" type="button" onClick={() => courtsQuery.refetch()}>{t('user.retry')}</button></div>}
          {!courtsQuery.isLoading && !courtsQuery.isError && courtsQuery.data?.length === 0 && <p className="panel-muted">{t('portal.noCourts')}</p>}
          <div className="court-list manager-list">
            {courtsQuery.data?.map((court) => (
              <div className="court-row" key={court.id}>
                <span>
                  <strong>{court.name}</strong>
                  <small>{court.slot_interval_minutes} {t('portal.minutes')} · {court.status}</small>
                </span>
                {canManageCourts && <span className="court-actions">
                  <button className="icon-button bordered" type="button" onClick={() => { setEditingCourtId(court.id); setCourtName(court.name); setSlotInterval(court.slot_interval_minutes) }} aria-label={t('portal.editCourt')}>
                    <Pencil size={16} />
                  </button>
                  <button className="icon-button bordered" type="button" onClick={() => { if (window.confirm(t('portal.archiveConfirm', { name: court.name }))) archiveCourt.mutate(court.id) }} aria-label={t('portal.archiveCourt')}>
                    <Trash2 size={16} />
                  </button>
                </span>}
              </div>
            ))}
          </div>
        </div>}
      </section>

      <section className="portal-panel two-column-panels" id="bookings">
        <div><p className="eyebrow">{t('portal.operations')}</p><h2>{t('portal.schedule')}</h2><p className="panel-muted">{t('portal.scheduleDescription')}</p><Link className="panel-link" to="/#bookings">{t('portal.openSchedule')} <ChevronRight size={16} /></Link></div>
        <div id="staff"><p className="eyebrow">{t('portal.access')}</p><h2>{t('portal.staff')}</h2><p className="panel-muted">{t('portal.staffDescription')}</p><button className="panel-link" type="button" disabled>{t('portal.inviteStaff')} <ChevronRight size={16} /></button></div>
      </section>

      <button className="portal-settings" type="button" disabled><Settings2 size={17} /> {t('portal.tenantSettings')}</button>
    </PortalShell>
  )
}
