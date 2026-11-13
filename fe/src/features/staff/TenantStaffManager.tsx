import { BuiltinRole } from '@/constants/domain'
import { MembershipStatus, TenantStatus } from '@/constants/domain'
import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useI18n } from '@/lib/use-i18n'
import { tenantApi } from '@/features/tenant/tenant.api'
import { venueApi } from '@/features/venue/venue.api'
import { tenantStaffApi, type TenantMember } from './staff.api'

export function TenantStaffManager() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const tenants = useQuery({ queryKey: ['tenants'], queryFn: tenantApi.list })
  const venues = useQuery({ queryKey: ['tenant-venues'], queryFn: venueApi.list })
  const [organizationId, setOrganizationId] = useState('')
  const [email, setEmail] = useState('')
  const [role, setRole] = useState<'tenant_staff' | 'venue_staff'>('tenant_staff')
  const activeTenants = tenants.data?.filter((tenant) => tenant.status === TenantStatus.Active) ?? []
  const selectedId = activeTenants.some((tenant) => tenant.id === organizationId) ? organizationId : activeTenants[0]?.id ?? ''
  const members = useQuery({ queryKey: ['tenant-staff', selectedId], queryFn: () => tenantStaffApi.list(selectedId), enabled: Boolean(selectedId) })
  const refresh = async () => { await queryClient.invalidateQueries({ queryKey: ['tenant-staff'] }) }
  const add = useMutation({ mutationFn: () => tenantStaffApi.add(selectedId, email.trim(), role), onSuccess: () => setEmail(''), onSettled: refresh })
  const remove = useMutation({ mutationFn: tenantStaffApi.remove, onSettled: refresh })
  const activate = useMutation({ mutationFn: (member: TenantMember) => tenantStaffApi.update(member, MembershipStatus.Active), onSettled: refresh })
  const assign = useMutation({ mutationFn: ({ venueId, userId, active }: { venueId: string; userId: string; active: boolean }) => tenantStaffApi.assignVenue(venueId, userId, active), onSettled: refresh })
  const busy = add.isPending || remove.isPending || activate.isPending || assign.isPending
  const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); if (selectedId) add.mutate() }

  return <section className="portal-panel resource-manager" id="staff-management">
    <h2>{t('member.title')}</h2><p className="panel-muted">{t('member.description')}</p>
    {tenants.isPending && <p role="status">{t('portal.loading')}</p>}
    {tenants.isSuccess && activeTenants.length === 0 && <p>{t('member.noActiveTenant')}</p>}
    {(tenants.isError || members.isError || venues.isError) && <div role="alert"><p>{t('crud.loadError')}</p><button className="secondary-button" type="button" onClick={() => { void tenants.refetch(); void venues.refetch(); if (selectedId) void members.refetch() }}>{t('user.retry')}</button></div>}
    {selectedId && <>
      <label>{t('venue.organization')} <select value={selectedId} onChange={(event) => setOrganizationId(event.target.value)}>{activeTenants.map((tenant) => <option key={tenant.id} value={tenant.id}>{tenant.name}</option>)}</select></label>
      <form className="resource-form" onSubmit={submit}><fieldset disabled={busy}>
        <label><span>{t('crud.email')}</span><input type="email" required value={email} onChange={(event) => setEmail(event.target.value)} /></label>
        <label><span>{t('crud.role')}</span><select value={role} onChange={(event) => setRole(event.target.value === BuiltinRole.VenueStaff ? 'venue_staff' : 'tenant_staff')}><option value="tenant_staff">{t('member.tenantStaff')}</option><option value="venue_staff">{t('member.venueStaff')}</option></select></label>
        <button className="primary-button" type="submit">{busy ? t('crud.busy') : t('crud.create')}</button>
      </fieldset></form>
      {(add.isError || remove.isError || activate.isError || assign.isError) && <p className="form-error" role="alert">{t('member.error')}</p>}
      {members.isPending && <p role="status">{t('portal.loading')}</p>}
      {members.isSuccess && members.data.length === 0 && <p>{t('crud.empty')}</p>}
      <ul className="resource-list">{members.data?.map((member) => <li key={member.user_id}>
        <div><strong>{member.full_name}</strong><p>{member.email}</p><p className="panel-muted">{t(member.role === BuiltinRole.TenantOwner ? 'member.owner' : member.role === BuiltinRole.TenantStaff ? 'member.tenantStaff' : 'member.venueStaff')} · {member.status === MembershipStatus.Active ? t('crud.active') : t('crud.suspended')}</p>
          {member.role === BuiltinRole.VenueStaff && member.status === MembershipStatus.Active && <div className="resource-actions">{venues.data?.filter((venue) => venue.organization_id === selectedId).map((venue) => <label key={venue.id}><input type="checkbox" checked={member.venue_ids?.includes(venue.id) ?? false} disabled={busy} onChange={(event) => assign.mutate({ venueId: venue.id, userId: member.user_id, active: event.target.checked })} /> {venue.name}</label>)}</div>}
        </div>
        {member.role !== BuiltinRole.TenantOwner && (member.status === MembershipStatus.Inactive
          ? <button className="secondary-button" type="button" disabled={busy} onClick={() => activate.mutate(member)}>{t('member.activate')}</button>
          : <button className="text-button" type="button" disabled={busy} onClick={() => { if (window.confirm(t('member.removeConfirm', { name: member.full_name }))) remove.mutate(member) }}>{t('member.remove')}</button>)}
      </li>)}</ul>
    </>}
  </section>
}
