import { createContext } from 'react'
import type { TranslationKey } from './locales/en'

export type Language = 'en' | 'th'
export type { TranslationKey }

export type LanguageContextValue = {
  language: Language
  setLanguage: (language: Language) => void
  t: (key: TranslationKey) => string
}

export const LanguageContext = createContext<LanguageContextValue | null>(null)
