import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useI18n } from '@/lib/use-i18n'
import type { TranslationKey } from '@/lib/i18n'

export interface CrudField {
  name: string
  label: TranslationKey
  type?: 'text' | 'email' | 'password' | 'select'
  options?: { value: string; label: string }[]
  required?: boolean
  maxLength?: number
  minLength?: number
  createOnly?: boolean
  editOnly?: boolean
}

interface CrudManagerProps<T extends { id: string }> {
  queryKey: string
  title: TranslationKey
  description?: TranslationKey
  fields: CrudField[]
  list: () => Promise<T[]>
  save: (id: string | null, values: Record<string, string>) => Promise<unknown>
  remove: (id: string) => Promise<unknown>
  values: (item: T) => Record<string, string>
  label: (item: T) => string
  summary: (item: T) => string
  canDelete?: (item: T) => boolean
}

export function CrudManager<T extends { id: string }>(props: CrudManagerProps<T>) {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [values, setValues] = useState<Record<string, string>>({})
  const query = useQuery({ queryKey: [props.queryKey], queryFn: props.list })
  const refresh = async () => {
    await Promise.all(['staff', 'roles', 'permissions'].map((key) => queryClient.invalidateQueries({ queryKey: [key] })))
  }
  const save = useMutation({
    mutationFn: () => props.save(editingId, values),
    onSuccess: async () => { setEditingId(null); setValues({}); await refresh() }
  })
  const remove = useMutation({ mutationFn: props.remove, onSuccess: refresh })
  const busy = save.isPending || remove.isPending
  const reset = () => { setEditingId(null); setValues({}); save.reset(); remove.reset() }
  const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); save.mutate() }

  return (
    <section className="portal-panel resource-manager" id={props.queryKey}>
      <h2>{t(props.title)}</h2>
      {props.description && <p className="panel-muted">{t(props.description)}</p>}
      <form onSubmit={submit} className="resource-form">
        <fieldset disabled={busy}>
          <legend>{editingId ? t('crud.edit') : t('crud.create')}</legend>
          {props.fields.filter((field) => editingId ? !field.createOnly : !field.editOnly).map((field) => (
            <label key={field.name}>
              <span>{t(field.label)}</span>
              {field.type === 'select' ? (
                <select value={values[field.name] ?? ''} required={field.required} onChange={(event) => setValues({ ...values, [field.name]: event.target.value })}>
                  <option value="" disabled>—</option>
                  {field.options?.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
                </select>
              ) : (
                <input type={field.type ?? 'text'} value={values[field.name] ?? ''} required={field.required} maxLength={field.maxLength} minLength={field.minLength} autoComplete={field.type === 'password' ? 'new-password' : undefined} onChange={(event) => setValues({ ...values, [field.name]: event.target.value })} />
              )}
            </label>
          ))}
          <div className="resource-actions">
            <button className="primary-button" type="submit">{save.isPending ? t('crud.busy') : t('crud.save')}</button>
            {editingId && <button className="secondary-button" type="button" onClick={reset}>{t('crud.cancel')}</button>}
          </div>
        </fieldset>
      </form>
      {save.isError && <p className="form-error" role="alert">{t('crud.saveError')}</p>}
      {remove.isError && <p className="form-error" role="alert">{t('crud.deleteError')}</p>}
      {query.isPending && <p role="status">{t('portal.loading')}</p>}
      {query.isError && <div role="alert"><p>{t('crud.loadError')}</p><button className="secondary-button" type="button" onClick={() => void query.refetch()}>{t('user.retry')}</button></div>}
      {query.isSuccess && query.data.length === 0 && <p>{t('crud.empty')}</p>}
      <ul className="resource-list">
        {query.data?.map((item) => (
          <li key={item.id}>
            <div><strong>{props.label(item)}</strong><p className="panel-muted">{props.summary(item)}</p></div>
            <div className="resource-actions">
              <button className="secondary-button" type="button" disabled={busy} onClick={() => { reset(); setEditingId(item.id); setValues(props.values(item)) }}>{t('crud.edit')}</button>
              <button className="text-button" type="button" disabled={busy || props.canDelete?.(item) === false} onClick={() => { if (window.confirm(t('crud.deleteConfirm', { name: props.label(item) }))) { reset(); remove.mutate(item.id) } }}>{t('crud.delete')}</button>
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}
