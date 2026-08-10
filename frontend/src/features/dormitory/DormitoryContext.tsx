import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import { useAuth } from '@/features/auth/AuthContext'
import { dormitoryService } from '@/features/dormitory/dormitoryService'
import type { ApiDormitory } from '@/types'

interface DormitoryContextValue {
  dormitories: ApiDormitory[]
  selectedDormitoryId: string | null
  isReady: boolean
  selectDormitory: (id: string) => void
}

const DormitoryContext = createContext<DormitoryContextValue | null>(null)

const SELECTED_DORMITORY_KEY = 'selected_dormitory_id'

export function DormitoryProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  const [dormitories, setDormitories] = useState<ApiDormitory[]>([])
  const [isReady, setIsReady] = useState(() => !isAuthenticated)
  const [selectedDormitoryId, setSelectedDormitoryId] = useState<string | null>(() =>
    localStorage.getItem(SELECTED_DORMITORY_KEY)
  )

  useEffect(() => {
    if (!isAuthenticated) {
      setDormitories([])
      setSelectedDormitoryId(null)
      setIsReady(true)
      return
    }

    let cancelled = false
    setIsReady(false)
    dormitoryService
      .mine()
      .then((list) => {
        if (cancelled) return
        setDormitories(list)

        const storedId = localStorage.getItem(SELECTED_DORMITORY_KEY)
        const stillValid = storedId && list.some((d) => d.id === storedId)
        if (!stillValid) {
          const fallback = list[0]?.id ?? null
          if (fallback) localStorage.setItem(SELECTED_DORMITORY_KEY, fallback)
          else localStorage.removeItem(SELECTED_DORMITORY_KEY)
          setSelectedDormitoryId(fallback)
        } else {
          setSelectedDormitoryId(storedId)
        }
        setIsReady(true)
      })
      .catch(() => {
        if (cancelled) return
        setDormitories([])
        setSelectedDormitoryId(null)
        localStorage.removeItem(SELECTED_DORMITORY_KEY)
        setIsReady(true)
      })

    return () => {
      cancelled = true
    }
  }, [isAuthenticated])

  const selectDormitory = useCallback((id: string) => {
    localStorage.setItem(SELECTED_DORMITORY_KEY, id)
    window.location.reload()
  }, [])

  return (
    <DormitoryContext.Provider value={{ dormitories, selectedDormitoryId, isReady, selectDormitory }}>
      {children}
    </DormitoryContext.Provider>
  )
}

export function useDormitory() {
  const ctx = useContext(DormitoryContext)
  if (!ctx) throw new Error('useDormitory must be used inside DormitoryProvider')
  return ctx
}
