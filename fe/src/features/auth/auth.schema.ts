import { z } from 'zod'

export const loginSchema = z.object({
  email: z.string().min(1, 'Vui lòng nhập email').email('Email không hợp lệ'),
  password: z.string().min(1, 'Vui lòng nhập mật khẩu')
})

export const registerSchema = z.object({
  full_name: z.string().trim().min(2, 'Họ tên phải có ít nhất 2 ký tự').max(120),
  email: z.string().min(1, 'Vui lòng nhập email').email('Email không hợp lệ'),
  phone: z.string().trim().max(20, 'Số điện thoại quá dài').optional(),
  password: z.string().min(8, 'Mật khẩu phải có ít nhất 8 ký tự').max(72)
})

export type LoginForm = z.infer<typeof loginSchema>
export type RegisterForm = z.infer<typeof registerSchema>
