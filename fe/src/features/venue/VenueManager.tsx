import { TenantStatus } from '@/constants/domain'
import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useI18n } from '@/lib/use-i18n'
import { tenantApi } from '@/features/tenant/tenant.api'
import type { Venue } from '@/features/catalog/catalog.api'
import { venueApi, type VenuePayload } from './venue.api'

const empty: VenuePayload = { name: '', slug: '', address: '', timezone: 'Asia/Ho_Chi_Minh' }

export function VenueManager() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const tenants = useQuery({ queryKey: ['tenants'], queryFn: tenantApi.list })
  const venues = useQuery({ queryKey: ['tenant-venues'], queryFn: venueApi.list })
  const activeTenants = tenants.data?.filter((tenant) => tenant.status === TenantStatus.Active) ?? []
  const [organizationId, setOrganizationId] = useState('')
  const [editing, setEditing] = useState<Venue | null>(null)
  const [payload, setPayload] = useState<VenuePayload>(empty)
  const selectedOrganizationId = editing?.organization_id ?? (activeTenants.some((tenant) => tenant.id === organizationId) ? organizationId : activeTenants[0]?.id ?? '')
  const refresh = async () => { await queryClient.invalidateQueries({ queryKey: ['tenant-venues'] }) }
  const save = useMutation({
    mutationFn: () => editing ? venueApi.update(editing.id, payload) : venueApi.create(selectedOrganizationId, payload),
    onSuccess: async () => { setEditing(null); setPayload(empty); await refresh() }
  })
  const archive = useMutation({ mutationFn: venueApi.archive, onSuccess: async () => { setEditing(null); setPayload(empty); await refresh() } })
  const busy = save.isPending || archive.isPending
  const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); if (selectedOrganizationId) save.mutate() }

  return <section className="portal-panel resource-manager" id="venue-management">
    <h2>{t('venue.title')}</h2>
    {(tenants.isPending || venues.isPending) && <p role="status">{t('portal.loading')}</p>}
    {(tenants.isError || venues.isError) && <div role="alert"><p>{t('crud.loadError')}</p><button className="secondary-button" type="button" onClick={() => { void tenants.refetch(); void venues.refetch() }}>{t('user.retry')}</button></div>}
    {tenants.isSuccess && activeTenants.length === 0 && <p role="status">{t('venue.approvalRequired')}</p>}
    {(activeTenants.length > 0 || editing) && <form className="resource-form" onSubmit={submit}>
      <fieldset disabled={busy}>
        <legend>{editing ? t('crud.edit') : t('venue.create')}</legend>
        <label><span>{t('venue.organization')}</span><select value={selectedOrganizationId} disabled={Boolean(editing)} onChange={(event) => setOrganizationId(event.target.value)}>
          {(editing ? tenants.data?.filter((tenant) => tenant.id === editing.organization_id) : activeTenants)?.map((tenant) => <option key={tenant.id} value={tenant.id}>{tenant.name}</option>)}
        </select></label>
        <label><span>{t('crud.name')}</span><input required maxLength={120} value={payload.name} onChange={(event) => setPayload({ ...payload, name: event.target.value })} /></label>
        <label><span>{t('crud.slug')}</span><input required maxLength={100} pattern="[a-z0-9]+(-[a-z0-9]+)*" value={payload.slug} onChange={(event) => setPayload({ ...payload, slug: event.target.value })} /></label>
        <label><span>{t('venue.address')}</span><input required maxLength={500} value={payload.address} onChange={(event) => setPayload({ ...payload, address: event.target.value })} /></label>
        <label><span>{t('venue.timezone')}</span><input required value={payload.timezone} onChange={(event) => setPayload({ ...payload, timezone: event.target.value })} /></label>
        <div className="resource-actions"><button className="primary-button" type="submit">{busy ? t('crud.busy') : t('crud.save')}</button>{editing && <button className="secondary-button" type="button" onClick={() => { setEditing(null); setPayload(empty); save.reset() }}>{t('crud.cancel')}</button>}</div>
      </fieldset>
    </form>}
    {(save.isError || archive.isError) && <p className="form-error" role="alert">{t('venue.error')}</p>}
    {venues.isSuccess && venues.data.length === 0 && <p>{t('crud.empty')}</p>}
    <ul className="resource-list">{venues.data?.map((venue) => <li key={venue.id}>
      <div><strong>{venue.name}</strong><p>{venue.address}</p><p className="panel-muted">{venue.slug} · {venue.status}</p></div>
      <div className="resource-actions"><button className="secondary-button" type="button" disabled={busy} onClick={() => { save.reset(); archive.reset(); setEditing(venue); setPayload({ name: venue.name, slug: venue.slug, address: venue.address, timezone: venue.timezone }) }}>{t('crud.edit')}</button><button className="text-button" type="button" disabled={busy} onClick={() => { if (window.confirm(t('venue.archiveConfirm', { name: venue.name }))) archive.mutate(venue.id) }}>{t('tenant.archive')}</button></div>
    </li>)}</ul>
  </section>
}
