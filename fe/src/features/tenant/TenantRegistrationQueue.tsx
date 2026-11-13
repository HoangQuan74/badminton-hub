import { TenantStatus } from '@/constants/domain'
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useI18n } from '@/lib/use-i18n'
import { tenantReviewApi } from './tenant-review.api'

export function TenantRegistrationQueue() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [status, setStatus] = useState<string>(TenantStatus.PendingVerification)
  const [notes, setNotes] = useState<Record<string, string>>({})
  const query = useQuery({ queryKey: ['tenant-registrations', status], queryFn: () => tenantReviewApi.list(status) })
  const review = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: 'approve' | 'reject' }) => tenantReviewApi.review(id, decision, notes[id] ?? ''),
    onSettled: async () => { await queryClient.invalidateQueries({ queryKey: ['tenant-registrations'] }) }
  })

  return <section className="portal-panel resource-manager" id="organizations">
    <h2>{t('review.title')}</h2>
    <label>{t('crud.status')} <select value={status} onChange={(event) => { setStatus(event.target.value); review.reset() }}>
      <option value={TenantStatus.PendingVerification}>{t('tenant.status.pending')}</option>
      <option value={TenantStatus.Active}>{t('tenant.status.active')}</option>
      <option value={TenantStatus.Rejected}>{t('tenant.status.rejected')}</option>
      <option value="">{t('review.all')}</option>
    </select></label>
    {query.isPending && <p role="status">{t('portal.loading')}</p>}
    {query.isError && <div role="alert"><p>{t('crud.loadError')}</p><button className="secondary-button" type="button" onClick={() => void query.refetch()}>{t('user.retry')}</button></div>}
    {query.isSuccess && query.data.length === 0 && <p>{t('review.empty')}</p>}
    {review.isError && <p role="alert" className="form-error">{t('review.error')}</p>}
    <div className="registration-grid">
      {query.data?.map((item) => <article className="portal-panel" key={item.organization.id}>
        <h3>{item.organization.name}</h3>
        <p>{item.organization.slug} · {item.organization.status}</p>
        <p>{item.contact_name}</p>
        <p><a href={`mailto:${item.contact_email}`}>{item.contact_email}</a></p>
        {item.contact_phone && <p><a href={`tel:${item.contact_phone.replace(/[^\d+]/g, '')}`}>{item.contact_phone}</a></p>}
        {item.note && <p>{item.note}</p>}
        {item.organization.status === TenantStatus.PendingVerification ? <>
          <label className="resource-note">{t('review.note')}<textarea maxLength={1000} value={notes[item.organization.id] ?? ''} disabled={review.isPending} onChange={(event) => setNotes({ ...notes, [item.organization.id]: event.target.value })} /></label>
          <div className="resource-actions">
            <button className="primary-button" type="button" disabled={review.isPending} onClick={() => { if (window.confirm(t('review.confirmApprove', { name: item.organization.name }))) review.mutate({ id: item.organization.id, decision: 'approve' }) }}>{t('review.approve')}</button>
            <button className="secondary-button" type="button" disabled={review.isPending} onClick={() => { if (window.confirm(t('review.confirmReject', { name: item.organization.name }))) review.mutate({ id: item.organization.id, decision: 'reject' }) }}>{t('review.reject')}</button>
          </div>
        </> : <p>{item.review_note}</p>}
      </article>)}
    </div>
  </section>
}
