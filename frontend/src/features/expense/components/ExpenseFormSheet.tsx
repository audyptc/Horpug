import type { FormEvent } from 'react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import { DormitorySearchSelect } from '@/features/dormitory/components/DormitorySearchSelect'
import type { ApiDormitory } from '@/features/dormitory/types'
import type { ExpenseCategory } from '../types'
import { EXPENSE_CATEGORIES } from '../utils'

const expenseCategoryLabelKeys: Record<ExpenseCategory, TranslationKey> = {
  maintenance: 'expenseCategoryMaintenance',
  utility: 'expenseCategoryUtility',
  salary: 'expenseCategorySalary',
  supplies: 'expenseCategorySupplies',
  other: 'expenseCategoryOther',
}

type ExpenseFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  dormitoryName: string
  onDormitorySelect: (dormitory: ApiDormitory) => void
  category: ExpenseCategory
  onCategoryChange: (category: ExpenseCategory) => void
  expenseDate: string
  onExpenseDateChange: (value: string) => void
  amount: string
  onAmountChange: (value: string) => void
  description: string
  onDescriptionChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function ExpenseFormSheet({
  open,
  onOpenChange,
  isEdit,
  dormitoryName,
  onDormitorySelect,
  category,
  onCategoryChange,
  expenseDate,
  onExpenseDateChange,
  amount,
  onAmountChange,
  description,
  onDescriptionChange,
  saving,
  error,
  onSubmit,
}: ExpenseFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('expenseFormEditTitle') : t('expenseFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('expenseFormEditDescription') : t('expenseFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <div className="flex flex-col gap-1.5 text-sm font-medium">
              {t('expenseFormDormitoryLabel')}
              <DormitorySearchSelect
                selectedLabel={dormitoryName}
                onSelectDormitory={onDormitorySelect}
                placeholder={t('expenseFormDormitoryPlaceholder')}
                searchPlaceholder={t('expenseFormDormitorySearchPlaceholder')}
                noResultsLabel={t('expenseFormDormitoryNoResults')}
                disabled={isEdit}
              />
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('expenseFormCategoryLabel')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={category}
                onChange={(event) => onCategoryChange(event.target.value as ExpenseCategory)}
              >
                {EXPENSE_CATEGORIES.map((value) => (
                  <option key={value} value={value}>
                    {t(expenseCategoryLabelKeys[value])}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('expenseFormDateLabel')}
              <input
                type="date"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={expenseDate}
                onChange={(event) => onExpenseDateChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('expenseFormAmountLabel')}
              <input
                type="number"
                min="0"
                step="0.01"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={amount}
                onChange={(event) => onAmountChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('expenseFormDescriptionLabel')}
              <textarea
                className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={description}
                onChange={(event) => onDescriptionChange(event.target.value)}
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('expenseFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('expenseSaving') : t('expenseFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
