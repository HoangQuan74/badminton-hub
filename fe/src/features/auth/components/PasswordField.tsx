import { useState, type InputHTMLAttributes } from 'react'
import { Eye, EyeOff } from 'lucide-react'

export function PasswordField(props: InputHTMLAttributes<HTMLInputElement>) {
  const [visible, setVisible] = useState(false)
  return (
    <div className="password-field">
      <input {...props} type={visible ? 'text' : 'password'} />
      <button
        className="icon-button"
        type="button"
        onClick={() => setVisible((value) => !value)}
        aria-label={visible ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
      >
        {visible ? <EyeOff size={18} /> : <Eye size={18} />}
      </button>
    </div>
  )
}
