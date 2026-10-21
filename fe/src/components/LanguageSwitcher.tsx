import { useI18n } from '@/lib/use-i18n'

export function LanguageSwitcher() {
  const { language, setLanguage, t } = useI18n()
  return (
    <label className="language-switcher">
      <span className="sr-only">{t('language.label')}</span>
      <select value={language} onChange={(event) => setLanguage(event.target.value as 'vi' | 'en')} aria-label={t('language.label')}>
        <option value="vi">VI</option>
        <option value="en">EN</option>
      </select>
    </label>
  )
}
