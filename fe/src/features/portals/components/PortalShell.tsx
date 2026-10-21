import type { LucideIcon } from 'lucide-react'
import { LogOut } from 'lucide-react'
import type { ReactNode } from 'react'
import { NavLink } from 'react-router'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { useI18n } from '@/lib/use-i18n'

interface PortalNavigationItem {
  label: string
  href: string
  icon: LucideIcon
}

interface PortalShellProps {
  portal: 'Tenant' | 'Admin'
  title: string
  navigation: PortalNavigationItem[]
  userName: string
  onLogout: () => void
  logoutPending: boolean
  children: ReactNode
}

export function PortalShell({ portal, title, navigation, userName, onLogout, logoutPending, children }: PortalShellProps) {
  const { t } = useI18n()
  const portalLabel = portal === 'Tenant' ? t('portal.tenant') : t('portal.admin')
  return (
    <div className="portal-shell">
      <aside className="portal-sidebar">
        <a className="brand" href="/"><span className="brand-mark">B</span><span>BadmintonHub</span></a>
        <div className="portal-label">{portalLabel}</div>
        <nav className="portal-nav" aria-label={portalLabel}>
          {navigation.map(({ label, href, icon: Icon }) => <NavLink key={href} to={href} end={href === `/tenant` || href === `/admin`}><Icon size={18} />{label}</NavLink>)}
        </nav>
        <div className="portal-sidebar-footer"><a href="/">← {t('portal.backToUser')}</a></div>
      </aside>
      <div className="portal-main">
        <header className="portal-header"><div><p className="eyebrow">{t('portal.workspace', { portal: portalLabel })}</p><h2>{title}</h2></div><div className="portal-user"><LanguageSwitcher /><span>{userName}</span><button className="icon-button bordered" type="button" onClick={onLogout} disabled={logoutPending} aria-label={t('portal.logout')}><LogOut size={18} /></button></div></header>
        <main className="portal-content">{children}</main>
      </div>
    </div>
  )
}
