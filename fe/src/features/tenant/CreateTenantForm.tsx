import { useState, type FormEvent } from 'react'
import axios from 'axios'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { portalUrls } from '@/app/portal'
import { useAuthStore } from '@/features/auth/auth.store'
import { getApiErrorMessage } from '@/lib/http'
import { useI18n } from '@/lib/use-i18n'
import { tenantApi, type TenantRegistrationPayload } from './tenant.api'

interface Attempt { key: string; payload: TenantRegistrationPayload }

function readAttempt(storageKey: string): Attempt | null {
  try {
    const value: unknown = JSON.parse(sessionStorage.getItem(storageKey) ?? 'null')
    if (typeof value !== 'object' || value === null || !('key' in value) || !('payload' in value)) return null
    const payload = value.payload
    if (typeof value.key !== 'string' || typeof payload !== 'object' || payload === null || !('name' in payload) || !('slug' in payload)) return null
    if (typeof payload.name !== 'string' || typeof payload.slug !== 'string') return null
    return { key: value.key, payload: {
 name: payload.name, slug: payload.slug,
 contact_name: 'contact_name' in payload && typeof payload.contact_name === 'string' ? payload.contact_name : '',
 contact_email: 'contact_email' in payload && typeof payload.contact_email === 'string' ? payload.contact_email : '',
 contact_phone: 'contact_phone' in payload && typeof payload.contact_phone === 'string' ? payload.contact_phone : '',
 note: 'note' in payload && typeof payload.note === 'string' ? payload.note : ''
 } }
  } catch { return null }
}

export function CreateTenantForm() {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const user = useAuthStore((state) => state.user)!
  const setUser = useAuthStore((state) => state.setUser)
  const storageKey = `tenant-create:${user.id}`
  const [attempt, setAttempt] = useState<Attempt | null>(() => readAttempt(storageKey))
  const [payload, setPayload] = useState<TenantRegistrationPayload>(() => attempt?.payload ?? { name: '', slug: '', contact_name: user.full_name, contact_email: user.email, contact_phone: user.phone, note: '' })
  const mutation = useMutation({
    mutationFn: (current: Attempt) => tenantApi.create(current.payload, current.key),
    onSuccess: async (result) => {
      sessionStorage.removeItem(storageKey)
      setAttempt(null)
      setUser(result.user)
      await queryClient.invalidateQueries({ queryKey: ['tenants'] })
    },
    onError: (error) => {
      // A definite client rejection permits editing; unknown outcomes keep the same key.
      const status = axios.isAxiosError(error) ? error.response?.status : undefined
      if (status && status >= 400 && status < 500 && status !== 408 && status !== 429) {
        sessionStorage.removeItem(storageKey)
        setAttempt(null)
      }
    }
  })

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const current = attempt ?? { key: crypto.randomUUID(), payload: { ...payload, name: payload.name.trim(), slug: payload.slug.trim().toLowerCase(), contact_name: payload.contact_name.trim(), contact_email: payload.contact_email.trim().toLowerCase(), contact_phone: payload.contact_phone.trim(), note: payload.note.trim() } }
    sessionStorage.setItem(storageKey, JSON.stringify(current))
    setAttempt(current)
    mutation.mutate(current)
  }

  if (mutation.isSuccess) {
    return <section className="portal-panel" role="status">
      <h2>{t('tenant.created', { name: mutation.data.organization.name })}</h2>
      <p>{t('tenant.pending')}</p>
      <a className="primary-button" href={`${portalUrls.tenant}/login`}>{t('tenant.login')}</a>
    </section>
  }

  return <section className="portal-panel resource-manager">
    <h2>{t('registration.title')}</h2>
    <p className="panel-muted">{t('registration.description')}</p>
    <form onSubmit={submit} className="resource-form">
      <fieldset disabled={mutation.isPending || attempt !== null}>
        <label><span>{t('crud.name')}</span><input required maxLength={120} value={payload.name} onChange={(event) => setPayload({ ...payload, name: event.target.value })} /></label>
        <label><span>{t('crud.slug')}</span><input required maxLength={100} pattern="[a-z0-9]+(-[a-z0-9]+)*" value={payload.slug} onChange={(event) => setPayload({ ...payload, slug: event.target.value })} /></label>
        <label><span>{t('registration.contactName')}</span><input required maxLength={120} autoComplete="name" value={payload.contact_name} onChange={(event) => setPayload({ ...payload, contact_name: event.target.value })} /></label>
        <label><span>{t('registration.contactEmail')}</span><input type="email" required maxLength={254} autoComplete="email" value={payload.contact_email} onChange={(event) => setPayload({ ...payload, contact_email: event.target.value })} /></label>
        <label><span>{t('registration.contactPhone')}</span><input type="tel" required maxLength={30} autoComplete="tel" value={payload.contact_phone} onChange={(event) => setPayload({ ...payload, contact_phone: event.target.value })} /></label>
        <label><span>{t('registration.note')}</span><textarea maxLength={1000} value={payload.note} onChange={(event) => setPayload({ ...payload, note: event.target.value })} /></label>
      </fieldset>
      {mutation.isError && <p className="form-error" role="alert">{getApiErrorMessage(mutation.error, t('tenant.error'))}</p>}
      {attempt && !mutation.isPending && <p role="status">{t('tenant.retry')}</p>}
      <button className="primary-button" type="submit" disabled={mutation.isPending}>{mutation.isPending ? t('crud.busy') : attempt ? t('tenant.retryButton') : t('tenant.create')}</button>
    </form>
  </section>
}
