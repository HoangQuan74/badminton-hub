import { z } from 'zod'
import type { TranslationKey } from '@/lib/i18n'

type Translate = (key: TranslationKey) => string

export function createAuthSchemas(t: Translate) {
  const loginSchema = z.object({
    email: z.string().min(1, t('validation.emailRequired')).email(t('validation.emailInvalid')),
    password: z.string().min(1, t('validation.passwordRequired'))
  })

  const registerSchema = z.object({
    full_name: z.string().trim().min(2, t('validation.nameLength')).max(120),
    email: z.string().min(1, t('validation.emailRequired')).email(t('validation.emailInvalid')),
    phone: z.string().trim().max(20, t('validation.phoneLength')).optional(),
    password: z.string().min(8, t('validation.passwordLength')).max(72)
  })

  return { loginSchema, registerSchema }
}

export type LoginForm = z.infer<ReturnType<typeof createAuthSchemas>['loginSchema']>
export type RegisterForm = z.infer<ReturnType<typeof createAuthSchemas>['registerSchema']>
