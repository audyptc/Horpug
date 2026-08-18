import { useEffect, useMemo, useState, type ReactNode } from 'react'
import en from './locales/en'
import th from './locales/th'
import { LanguageContext, type Language, type LanguageContextValue } from './language-context'

const translations = { en, th } as const

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<Language>(
    () => (localStorage.getItem('language') as Language | null) ?? 'en'
  )

  useEffect(() => {
    localStorage.setItem('language', language)
    document.documentElement.lang = language
  }, [language])

  const value = useMemo<LanguageContextValue>(
    () => ({
      language,
      setLanguage,
      t: (key) => translations[language][key],
    }),
    [language]
  )

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>
}
