export type ExpenseCategory = 'repair' | 'utilities' | 'supplies' | 'salary' | 'other'

export interface ApiExpense {
  id: string
  expense_date: string
  category: ExpenseCategory
  description: string
  amount: number
  note: string
  created_at: string
  updated_at: string
}

export interface CreateExpensePayload {
  expense_date: string
  category: ExpenseCategory
  description: string
  amount: number
  note: string
}

export interface UpdateExpensePayload {
  expense_date: string
  category: ExpenseCategory
  description: string
  amount: number
  note: string
}
