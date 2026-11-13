import { Link } from 'react-router'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { useI18n } from '@/lib/use-i18n'
import { CreateTenantForm } from '../CreateTenantForm'

export function TenantRegistrationPage() {
  const { t } = useI18n()
  return <main className="content-section">
    <div className="resource-actions"><Link to="/">{t('registration.back')}</Link><LanguageSwitcher /></div>
    <CreateTenantForm />
  </main>
}
