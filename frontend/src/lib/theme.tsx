import React, { createContext, useContext, useEffect, useState } from 'react'

type Theme = 'light' | 'dark'
export type SidebarColor = 'slate' | 'blue' | 'purple' | 'green' | 'light'

export const SIDEBAR_COLORS: { value: SidebarColor; label: string; bg: string }[] = [
  { value: 'slate',  label: 'Slate',  bg: 'hsl(240 5.9% 10%)' },
  { value: 'blue',   label: 'Blue',   bg: 'hsl(221 83% 16%)' },
  { value: 'purple', label: 'Purple', bg: 'hsl(262 60% 16%)' },
  { value: 'green',  label: 'Green',  bg: 'hsl(158 64% 12%)' },
  { value: 'light',  label: 'Light',  bg: 'hsl(0 0% 98%)' },
]

interface ThemeContextValue {
  theme: Theme
  toggleTheme: () => void
  setTheme: (t: Theme) => void
  sidebarColor: SidebarColor
  setSidebarColor: (c: SidebarColor) => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() => {
    const stored = localStorage.getItem('theme') as Theme | null
    if (stored === 'dark' || stored === 'light') return stored
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  })

  const [sidebarColor, setSidebarColorState] = useState<SidebarColor>(() => {
    return (localStorage.getItem('sidebarColor') as SidebarColor) ?? 'slate'
  })

  useEffect(() => {
    const root = document.documentElement
    root.classList.toggle('dark', theme === 'dark')
    localStorage.setItem('theme', theme)
  }, [theme])

  useEffect(() => {
    document.documentElement.setAttribute('data-sidebar-color', sidebarColor)
    localStorage.setItem('sidebarColor', sidebarColor)
  }, [sidebarColor])

  function setTheme(t: Theme) { setThemeState(t) }
  function toggleTheme() { setThemeState((p) => (p === 'dark' ? 'light' : 'dark')) }
  function setSidebarColor(c: SidebarColor) { setSidebarColorState(c) }

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, setTheme, sidebarColor, setSidebarColor }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used inside ThemeProvider')
  return ctx
}
