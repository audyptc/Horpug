import { Plus, X } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import type { PresetRow } from './utils'

type DeductionPresetsEditorProps = {
  rows: PresetRow[]
  onChange: (rows: PresetRow[]) => void
  loading?: boolean
}

// Edits a dormitory's ready-made move-out deductions (e.g. cleaning 500),
// offered as one-tap buttons on the move-out screen.
export function DeductionPresetsEditor({ rows, onChange, loading = false }: DeductionPresetsEditorProps) {
  const { t } = useLanguage()

  function update(key: number, patch: Partial<PresetRow>) {
    onChange(rows.map((row) => (row.key === key ? { ...row, ...patch } : row)))
  }

  function add() {
    const nextKey = rows.reduce((max, row) => Math.max(max, row.key), 0) + 1
    onChange([...rows, { key: nextKey, name: '', amount: '' }])
  }

  return (
    <div className="flex flex-col gap-1.5 text-sm font-medium">
      <div className="flex items-center justify-between gap-2">
        <span>{t('dormitoryFormPresetsLabel')}</span>
        <Button type="button" size="sm" variant="outline" onClick={add} disabled={loading}>
          <Plus />
          {t('dormitoryFormPresetsAdd')}
        </Button>
      </div>
      <span className="text-xs font-normal text-muted-foreground">{t('dormitoryFormPresetsHint')}</span>
      {loading ? (
        <p className="metric-detail">{t('loading')}</p>
      ) : (
        rows.map((row) => (
          <div key={row.key} className="flex items-center gap-1.5">
            <input
              type="text"
              className="h-9 min-w-0 flex-1 rounded-md border border-input bg-transparent px-2.5 text-sm font-normal"
              placeholder={t('dormitoryFormPresetsName')}
              value={row.name}
              onChange={(event) => update(row.key, { name: event.target.value })}
            />
            <input
              type="number"
              min="0"
              step="0.01"
              className="h-9 w-28 shrink-0 rounded-md border border-input bg-transparent px-2.5 text-right text-sm font-normal"
              placeholder={t('dormitoryFormPresetsAmount')}
              value={row.amount}
              onChange={(event) => update(row.key, { amount: event.target.value })}
            />
            <Button
              type="button"
              size="icon"
              variant="ghost"
              className="h-8 w-8 shrink-0 text-muted-foreground"
              title={t('moveOutRemoveDeduction')}
              aria-label={t('moveOutRemoveDeduction')}
              onClick={() => onChange(rows.filter((r) => r.key !== row.key))}
            >
              <X className="size-4" />
            </Button>
          </div>
        ))
      )}
    </div>
  )
}
