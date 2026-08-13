export type ExpenseCategory = 'maintenance' | 'utility' | 'salary' | 'supplies' | 'other'

export type ApiExpense = {
  id: string
  dormitory_id: string
  dormitory_name?: string
  category: ExpenseCategory
  expense_date: string
  amount: number
  description: string
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}
