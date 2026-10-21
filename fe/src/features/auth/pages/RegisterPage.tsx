import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { ArrowRight } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, Navigate, useNavigate } from 'react-router'
import { getApiErrorMessage } from '@/lib/http'
import { authApi } from '../auth.api'
import { createAuthSchemas, type RegisterForm } from '../auth.schema'
import { useAuthStore } from '../auth.store'
import { AuthShell } from '../components/AuthShell'
import { PasswordField } from '../components/PasswordField'
import { useI18n } from '@/lib/use-i18n'

export function RegisterPage() {
  const { t } = useI18n()
  const user = useAuthStore((state) => state.user)
  const setUser = useAuthStore((state) => state.setUser)
  const navigate = useNavigate()
  const [serverError, setServerError] = useState('')
  const { registerSchema } = useMemo(() => createAuthSchemas(t), [t])
  const form = useForm<RegisterForm>({ resolver: zodResolver(registerSchema), defaultValues: { full_name: '', email: '', phone: '', password: '' } })
  const mutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: ({ user: registeredUser }) => { setUser(registeredUser); navigate('/', { replace: true }) },
    onError: (error) => setServerError(getApiErrorMessage(error, t('register.error')))
  })

  if (user) return <Navigate to="/" replace />

  return (
    <AuthShell>
      <div className="auth-form-wrap register-form-wrap">
        <p className="eyebrow">{t('register.eyebrow')}</p>
        <h2>{t('register.title')}</h2>
        <p className="form-intro">{t('register.intro')}</p>
        <form className="auth-form" onSubmit={form.handleSubmit((values) => { setServerError(''); mutation.mutate({ ...values, phone: values.phone || undefined }) })} noValidate>
          <label><span>{t('register.fullName')}</span><input autoComplete="name" placeholder={t('register.fullNamePlaceholder')} {...form.register('full_name')} />{form.formState.errors.full_name && <small className="field-error">{form.formState.errors.full_name.message}</small>}</label>
          <label><span>{t('login.email')}</span><input type="email" autoComplete="email" placeholder={t('login.emailPlaceholder')} {...form.register('email')} />{form.formState.errors.email && <small className="field-error">{form.formState.errors.email.message}</small>}</label>
          <label><span>{t('register.phone')} <em>{t('register.optional')}</em></span><input type="tel" autoComplete="tel" placeholder={t('register.phonePlaceholder')} {...form.register('phone')} />{form.formState.errors.phone && <small className="field-error">{form.formState.errors.phone.message}</small>}</label>
          <label><span>{t('login.password')}</span><PasswordField autoComplete="new-password" placeholder={t('register.passwordPlaceholder')} {...form.register('password')} />{form.formState.errors.password && <small className="field-error">{form.formState.errors.password.message}</small>}</label>
          {serverError && <div className="form-alert" role="alert">{serverError}</div>}
          <button className="primary-button" type="submit" disabled={mutation.isPending}>{mutation.isPending ? 'Đang tạo tài khoản…' : <>Tạo tài khoản <ArrowRight size={18} /></>}</button>
        </form>
        <p className="auth-switch">{t('register.hasAccount')} <Link to="/login">{t('register.login')}</Link></p>
      </div>
    </AuthShell>
  )
}
