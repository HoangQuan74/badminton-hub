import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { ArrowRight } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, Navigate, useLocation, useNavigate } from 'react-router'
import { getApiErrorMessage } from '@/lib/http'
import { authApi } from '../auth.api'
import { createAuthSchemas, type LoginForm } from '../auth.schema'
import { useAuthStore } from '../auth.store'
import { AuthShell } from '../components/AuthShell'
import { PasswordField } from '../components/PasswordField'
import { useI18n } from '@/lib/use-i18n'

export function LoginPage() {
  const { t } = useI18n()
  const user = useAuthStore((state) => state.user)
  const setUser = useAuthStore((state) => state.setUser)
  const navigate = useNavigate()
  const location = useLocation()
  const [serverError, setServerError] = useState('')
  const { loginSchema } = useMemo(() => createAuthSchemas(t), [t])
  const form = useForm<LoginForm>({ resolver: zodResolver(loginSchema), defaultValues: { email: '', password: '' } })
  const mutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: ({ user: signedInUser }) => {
      setUser(signedInUser)
      const destination = (location.state as { from?: string } | null)?.from || '/'
      navigate(destination, { replace: true })
    },
    onError: (error) => setServerError(getApiErrorMessage(error, t('login.error')))
  })

  if (user) return <Navigate to="/" replace />

  return (
    <AuthShell>
      <div className="auth-form-wrap">
        <p className="eyebrow">{t('login.eyebrow')}</p>
        <h2>{t('login.title')}</h2>
        <p className="form-intro">{t('login.intro')}</p>

        <form className="auth-form" onSubmit={form.handleSubmit((values) => { setServerError(''); mutation.mutate(values) })} noValidate>
          <label>
            <span>{t('login.email')}</span>
            <input type="email" autoComplete="email" placeholder={t('login.emailPlaceholder')} {...form.register('email')} />
            {form.formState.errors.email && <small className="field-error">{form.formState.errors.email.message}</small>}
          </label>
          <label>
            <span>{t('login.password')}</span>
            <PasswordField autoComplete="current-password" placeholder={t('login.passwordPlaceholder')} {...form.register('password')} />
            {form.formState.errors.password && <small className="field-error">{form.formState.errors.password.message}</small>}
          </label>
          <div className="form-row"><label className="check-label"><input type="checkbox" /> {t('login.remember')}</label><button className="text-button" type="button">{t('login.forgot')}</button></div>
          {serverError && <div className="form-alert" role="alert">{serverError}</div>}
          <button className="primary-button" type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? t('login.submitting') : <>{t('login.submit')} <ArrowRight size={18} /></>}
          </button>
        </form>
        <p className="auth-switch">{t('login.noAccount')} <Link to="/register">{t('login.register')}</Link></p>
      </div>
    </AuthShell>
  )
}
