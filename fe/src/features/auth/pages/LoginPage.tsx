import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { ArrowRight } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, Navigate, useLocation, useNavigate } from 'react-router'
import { getApiErrorMessage } from '@/lib/http'
import { authApi } from '../auth.api'
import { loginSchema, type LoginForm } from '../auth.schema'
import { useAuthStore } from '../auth.store'
import { AuthShell } from '../components/AuthShell'
import { PasswordField } from '../components/PasswordField'

export function LoginPage() {
  const user = useAuthStore((state) => state.user)
  const setUser = useAuthStore((state) => state.setUser)
  const navigate = useNavigate()
  const location = useLocation()
  const [serverError, setServerError] = useState('')
  const form = useForm<LoginForm>({ resolver: zodResolver(loginSchema), defaultValues: { email: '', password: '' } })
  const mutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: ({ user: signedInUser }) => {
      setUser(signedInUser)
      const destination = (location.state as { from?: string } | null)?.from || '/'
      navigate(destination, { replace: true })
    },
    onError: (error) => setServerError(getApiErrorMessage(error, 'Không thể đăng nhập. Vui lòng thử lại.'))
  })

  if (user) return <Navigate to="/" replace />

  return (
    <AuthShell>
      <div className="auth-form-wrap">
        <p className="eyebrow">Chào mừng trở lại</p>
        <h2>Đăng nhập tài khoản</h2>
        <p className="form-intro">Tiếp tục để tìm sân và quản lý lịch chơi của bạn.</p>

        <form className="auth-form" onSubmit={form.handleSubmit((values) => { setServerError(''); mutation.mutate(values) })} noValidate>
          <label>
            <span>Email</span>
            <input type="email" autoComplete="email" placeholder="ban@example.com" {...form.register('email')} />
            {form.formState.errors.email && <small className="field-error">{form.formState.errors.email.message}</small>}
          </label>
          <label>
            <span>Mật khẩu</span>
            <PasswordField autoComplete="current-password" placeholder="Nhập mật khẩu" {...form.register('password')} />
            {form.formState.errors.password && <small className="field-error">{form.formState.errors.password.message}</small>}
          </label>
          <div className="form-row"><label className="check-label"><input type="checkbox" /> Ghi nhớ tôi</label><button className="text-button" type="button">Quên mật khẩu?</button></div>
          {serverError && <div className="form-alert" role="alert">{serverError}</div>}
          <button className="primary-button" type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? 'Đang đăng nhập…' : <>Đăng nhập <ArrowRight size={18} /></>}
          </button>
        </form>
        <p className="auth-switch">Chưa có tài khoản? <Link to="/register">Đăng ký miễn phí</Link></p>
      </div>
    </AuthShell>
  )
}
