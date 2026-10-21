import { useState, type InputHTMLAttributes } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import { useI18n } from '@/lib/use-i18n'

export function PasswordField(props: InputHTMLAttributes<HTMLInputElement>) {
  const [visible, setVisible] = useState(false)
  const { t } = useI18n()
  return (
    <div className="password-field">
      <input {...props} type={visible ? 'text' : 'password'} />
      <button
        className="icon-button"
        type="button"
        onClick={() => setVisible((value) => !value)}
        aria-label={visible ? t('password.hide') : t('password.show')}
      >
        {visible ? <EyeOff size={18} /> : <Eye size={18} />}
      </button>
    </div>
  )
}
