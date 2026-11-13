import { BuiltinRole } from '@/constants/domain'
import { TenantStatus } from '@/constants/domain'
import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useI18n } from '@/lib/use-i18n'
import type { TranslationKey } from '@/lib/i18n'
import { getApiErrorMessage } from '@/lib/http'
import { useAuthStore } from '@/features/auth/auth.store'
import { tenantApi, type Tenant, type TenantPayload } from './tenant.api'

const statusKeys: Record<Tenant['status'], TranslationKey> = {
  PENDING_VERIFICATION: 'tenant.status.pending', ACTIVE: 'tenant.status.active', SUSPENDED: 'tenant.status.suspended', REJECTED: 'tenant.status.rejected', ARCHIVED: 'tenant.status.archived'
}

export function TenantManager() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const user = useAuthStore((state) => state.user)!
  const canManage = user.roles?.includes(BuiltinRole.TenantOwner) ?? false
  const query = useQuery({ queryKey: ['tenants'], queryFn: tenantApi.list })
  const [editing, setEditing] = useState<Tenant | null>(null)
  const [payload, setPayload] = useState<TenantPayload>({ name: '', slug: '' })
  const refresh = async () => { await queryClient.invalidateQueries({ queryKey: ['tenants'] }) }
  const save = useMutation({
    mutationFn: (id: string) => tenantApi.update(id, payload),
    onSuccess: async () => { setEditing(null); await refresh() }
  })
  const archive = useMutation({ mutationFn: tenantApi.archive, onSuccess: refresh })
  const busy = save.isPending || archive.isPending
  const error = save.error ?? archive.error
  const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); if (editing) save.mutate(editing.id) }

  return <section className="portal-panel resource-manager" id="tenants">
    <h2>{t('tenant.title')}</h2>
    {query.isPending && <p role="status">{t('portal.loading')}</p>}
    {query.isError && <div role="alert"><p>{t('crud.loadError')}</p><button type="button" className="secondary-button" onClick={() => void query.refetch()}>{t('user.retry')}</button></div>}
    {query.isSuccess && query.data.length === 0 && <p>{t('crud.empty')}</p>}
    {error && <p className="form-error" role="alert">{getApiErrorMessage(error, t('tenant.error'))}</p>}
    <ul className="resource-list">
      {query.data?.map((tenant) => <li key={tenant.id}>
        <div><strong>{tenant.name}</strong><p className="panel-muted">{tenant.slug} · {t(statusKeys[tenant.status])}</p></div>
        {canManage && <div className="resource-actions">
          <button className="secondary-button" type="button" disabled={busy || (tenant.status !== TenantStatus.PendingVerification && tenant.status !== TenantStatus.Active)} onClick={() => { save.reset(); archive.reset(); setEditing(tenant); setPayload({ name: tenant.name, slug: tenant.slug }) }}>{t('crud.edit')}</button>
          <button className="text-button" type="button" disabled={busy || tenant.status !== TenantStatus.PendingVerification} onClick={() => { if (window.confirm(t('tenant.archiveConfirm', { name: tenant.name }))) { setEditing(null); archive.mutate(tenant.id) } }}>{t('tenant.archive')}</button>
        </div>}
      </li>)}
    </ul>
    {editing && <form onSubmit={submit} className="resource-form">
      <fieldset disabled={busy}>
        <legend>{t('crud.edit')}</legend>
        <label><span>{t('crud.name')}</span><input required maxLength={120} value={payload.name} onChange={(event) => setPayload({ ...payload, name: event.target.value })} /></label>
        <label><span>{t('crud.slug')}</span><input required maxLength={100} pattern="[a-z0-9]+(-[a-z0-9]+)*" value={payload.slug} onChange={(event) => setPayload({ ...payload, slug: event.target.value })} /></label>
        <div className="resource-actions"><button className="primary-button" type="submit">{t('crud.save')}</button><button className="secondary-button" type="button" onClick={() => setEditing(null)}>{t('crud.cancel')}</button></div>
      </fieldset>
    </form>}
  </section>
}
