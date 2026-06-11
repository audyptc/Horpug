import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ApiExpense, ExpenseCategory } from '@/types'

type ExpenseForm = {
  expense_date: string
  category: ExpenseCategory
  description: string
  amount: string
  note: string
}

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingExpense: ApiExpense | null
  form: ExpenseForm
  onFormChange: (form: ExpenseForm) => void
  onSave: () => void
  saving: boolean
}

export function ExpenseDialog({
  open,
  onOpenChange,
  editingExpense,
  form,
  onFormChange,
  onSave,
  saving,
}: Props) {
  const { t } = useTranslation()

  const isSaveDisabled = saving || !form.expense_date || !form.description || !form.amount

  function set(patch: Partial<ExpenseForm>) {
    onFormChange({ ...form, ...patch })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {editingExpense ? t('expenses.editExpense') : t('expenses.createExpense')}
          </DialogTitle>
          <DialogDescription>
            {editingExpense ? t('expenses.editDesc') : t('expenses.createDesc')}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>{t('expenses.expenseDate')} *</Label>
              <Input
                type="date"
                value={form.expense_date}
                onChange={(e) => set({ expense_date: e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <Label>{t('expenses.category')} *</Label>
              <Select
                value={form.category}
                onValueChange={(v) => set({ category: v as ExpenseCategory })}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="repair">{t('expenses.categories.repair')}</SelectItem>
                  <SelectItem value="utilities">{t('expenses.categories.utilities')}</SelectItem>
                  <SelectItem value="supplies">{t('expenses.categories.supplies')}</SelectItem>
                  <SelectItem value="salary">{t('expenses.categories.salary')}</SelectItem>
                  <SelectItem value="other">{t('expenses.categories.other')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>{t('expenses.description')} *</Label>
            <Input
              placeholder={t('expenses.descriptionPlaceholder')}
              value={form.description}
              onChange={(e) => set({ description: e.target.value })}
            />
          </div>

          <div className="space-y-1.5">
            <Label>{t('expenses.amount')} *</Label>
            <Input
              type="number"
              min="0"
              step="0.01"
              placeholder="0.00"
              value={form.amount}
              onChange={(e) => set({ amount: e.target.value })}
            />
          </div>

          <div className="space-y-1.5">
            <Label>{t('expenses.note')}</Label>
            <textarea
              rows={2}
              placeholder={t('expenses.notePlaceholder')}
              value={form.note}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set({ note: e.target.value })}
              className="flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onSave} disabled={isSaveDisabled}>
            {saving
              ? t('common.loading')
              : editingExpense
                ? t('expenses.saveChanges')
                : t('expenses.createExpense')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
