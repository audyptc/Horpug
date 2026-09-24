export type PresetRow = { key: number; name: string; amount: string }

export function isCompletePreset(row: PresetRow): boolean {
  const amount = Number(row.amount)
  return row.name.trim() !== '' && Number.isFinite(amount) && amount > 0
}

export function formatBaht(value: number): string {
  return value.toLocaleString('th-TH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
