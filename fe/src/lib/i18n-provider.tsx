import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { I18nContext, LANGUAGE_STORAGE_KEY, messages, interpolate, type I18nContextValue, type Language } from './i18n'

function getInitialLanguage(): Language {
  if (typeof window === 'undefined') return 'vi'
  const stored = window.localStorage.getItem(LANGUAGE_STORAGE_KEY)
  if (stored === 'vi' || stored === 'en') return stored
  return 'vi'
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<Language>(getInitialLanguage)

  useEffect(() => {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, language)
    document.documentElement.lang = language
  }, [language])

  const value = useMemo<I18nContextValue>(() => ({
    language,
    setLanguage,
    t: (key, variables) => interpolate(messages[language][key] ?? messages.vi[key], variables)
  }), [language])

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}
