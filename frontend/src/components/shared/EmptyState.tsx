import { SearchX } from 'lucide-react'

interface EmptyStateProps {
  message: string
  onClear?: () => void
  clearLabel?: string
}

export function EmptyState({ message, onClear, clearLabel }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3 text-muted-foreground">
      <div className="p-4 rounded-full bg-muted">
        <SearchX className="w-6 h-6" />
      </div>
      <p className="text-sm font-medium">{message}</p>
      {onClear && clearLabel && (
        <button
          type="button"
          onClick={onClear}
          className="text-xs text-primary hover:underline"
        >
          {clearLabel}
        </button>
      )}
    </div>
  )
}
