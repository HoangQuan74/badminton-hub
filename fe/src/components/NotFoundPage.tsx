import { useI18n } from '@/lib/use-i18n'

export function NotFoundPage() {
  const { t } = useI18n()
  return <div className="not-found"><h1>404</h1><p>{t('notFound.title')}</p><a href="/">{t('notFound.back')}</a></div>
}
