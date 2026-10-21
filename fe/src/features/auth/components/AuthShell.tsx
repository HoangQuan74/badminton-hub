import type { ReactNode } from 'react'
import { Link } from 'react-router'
import { CalendarCheck, MapPin, ShieldCheck } from 'lucide-react'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { useI18n } from '@/lib/use-i18n'

export function AuthShell({ children }: { children: ReactNode }) {
  const { t } = useI18n()
  return (
    <main className="auth-shell">
      <section className="auth-story" aria-label={t('auth.storyLabel')}>
        <Link className="brand brand-on-dark" to="/">
          <span className="brand-mark">B</span>
          <span>BadmintonHub</span>
        </Link>
        <div className="auth-story-content">
          <p className="eyebrow light">{t('auth.eyebrow')}</p>
          <h1>{t('auth.title')}</h1>
          <p>{t('auth.description')}</p>
          <ul className="auth-benefits">
            <li><MapPin size={20} /> {t('auth.benefit.discovery')}</li>
            <li><CalendarCheck size={20} /> {t('auth.benefit.booking')}</li>
            <li><ShieldCheck size={20} /> {t('auth.benefit.security')}</li>
          </ul>
        </div>
        <p className="auth-quote">{t('auth.quote')}</p>
      </section>
      <section className="auth-panel"><div className="auth-language"><LanguageSwitcher /></div>{children}</section>
    </main>
  )
}
