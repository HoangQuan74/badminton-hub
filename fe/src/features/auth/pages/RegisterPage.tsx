import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { ArrowRight } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, Navigate, useNavigate } from 'react-router'
import { getApiErrorMessage } from '@/lib/http'
import { authApi } from '../auth.api'
import { registerSchema, type RegisterForm } from '../auth.schema'
import { useAuthStore } from '../auth.store'
import { AuthShell } from '../components/AuthShell'
import { PasswordField } from '../components/PasswordField'

export function RegisterPage() {
  const user = useAuthStore((state) => state.user)
  const setUser = useAuthStore((state) => state.setUser)
  const navigate = useNavigate()
  const [serverError, setServerError] = useState('')
  const form = useForm<RegisterForm>({ resolver: zodResolver(registerSchema), defaultValues: { full_name: '', email: '', phone: '', password: '' } })
  const mutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: ({ user: registeredUser }) => { setUser(registeredUser); navigate('/', { replace: true }) },
    onError: (error) => setServerError(getApiErrorMessage(error, 'Không thể tạo tài khoản. Vui lòng thử lại.'))
  })

  if (user) return <Navigate to="/" replace />

  return (
    <AuthShell>
      <div className="auth-form-wrap register-form-wrap">
        <p className="eyebrow">Bắt đầu cùng BadmintonHub</p>
        <h2>Tạo tài khoản</h2>
        <p className="form-intro">Lưu sân yêu thích và đặt lịch nhanh hơn cho lần chơi tiếp theo.</p>
        <form className="auth-form" onSubmit={form.handleSubmit((values) => { setServerError(''); mutation.mutate({ ...values, phone: values.phone || undefined }) })} noValidate>
          <label><span>Họ và tên</span><input autoComplete="name" placeholder="Nguyễn Hoàng Quân" {...form.register('full_name')} />{form.formState.errors.full_name && <small className="field-error">{form.formState.errors.full_name.message}</small>}</label>
          <label><span>Email</span><input type="email" autoComplete="email" placeholder="ban@example.com" {...form.register('email')} />{form.formState.errors.email && <small className="field-error">{form.formState.errors.email.message}</small>}</label>
          <label><span>Số điện thoại <em>(không bắt buộc)</em></span><input type="tel" autoComplete="tel" placeholder="09xx xxx xxx" {...form.register('phone')} />{form.formState.errors.phone && <small className="field-error">{form.formState.errors.phone.message}</small>}</label>
          <label><span>Mật khẩu</span><PasswordField autoComplete="new-password" placeholder="Tối thiểu 8 ký tự" {...form.register('password')} />{form.formState.errors.password && <small className="field-error">{form.formState.errors.password.message}</small>}</label>
          {serverError && <div className="form-alert" role="alert">{serverError}</div>}
          <button className="primary-button" type="submit" disabled={mutation.isPending}>{mutation.isPending ? 'Đang tạo tài khoản…' : <>Tạo tài khoản <ArrowRight size={18} /></>}</button>
        </form>
        <p className="auth-switch">Đã có tài khoản? <Link to="/login">Đăng nhập</Link></p>
      </div>
    </AuthShell>
  )
}
