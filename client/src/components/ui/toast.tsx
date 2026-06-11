import { createContext, useContext, useState, useCallback, useMemo } from 'react'
import { createPortal } from 'react-dom'
import { X, CheckCircle2, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

type ToastVariant = 'success' | 'error'

type Toast = {
  id: number
  message: string
  variant: ToastVariant
}

type ToastContextType = {
  success: (message: string) => void
  error: (message: string) => void
}

const ToastContext = createContext<ToastContextType | null>(null)

let idCounter = 0

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const add = useCallback((message: string, variant: ToastVariant) => {
    const id = ++idCounter
    setToasts((prev) => [...prev, { id, message, variant }])
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id))
    }, 3500)
  }, [])

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const ctx = useMemo<ToastContextType>(
    () => ({
      success: (msg) => add(msg, 'success'),
      error: (msg) => add(msg, 'error'),
    }),
    [add]
  )

  return (
    <ToastContext.Provider value={ctx}>
      {children}
      {createPortal(
        <div
          aria-live="polite"
          className="fixed top-4 right-4 z-9999 flex flex-col gap-2 w-80 max-w-[calc(100vw-2rem)] pointer-events-none"
        >
          {toasts.map((t) => (
            <div
              key={t.id}
              className={cn(
                'flex items-start gap-3 rounded-lg border px-4 py-3 shadow-lg text-sm pointer-events-auto',
                t.variant === 'success'
                  ? 'bg-green-50 border-green-200 text-green-800 dark:bg-green-950 dark:border-green-800 dark:text-green-200'
                  : 'bg-red-50 border-red-200 text-red-800 dark:bg-red-950 dark:border-red-800 dark:text-red-200'
              )}
            >
              {t.variant === 'success' ? (
                <CheckCircle2 className="w-4 h-4 mt-0.5 shrink-0" />
              ) : (
                <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
              )}
              <span className="flex-1">{t.message}</span>
              <button
                type="button"
                onClick={() => remove(t.id)}
                className="shrink-0 opacity-60 hover:opacity-100"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>,
        document.body
      )}
    </ToastContext.Provider>
  )
}

export function useToast(): ToastContextType {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}
