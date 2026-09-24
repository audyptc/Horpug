export type ReportAmount = { key: string; amount: number }

export type ApiMonthlyReport = {
  year: number
  month: number
  income: { count: number; total: number; by_method: ReportAmount[] }
  billing: {
    invoice_count: number
    billed: number
    collected: number
    outstanding: number
    by_item_type: ReportAmount[]
  }
  expenses: { total: number; by_category: ReportAmount[] }
  net: number
  arrears: { count: number; amount: number }
  occupancy: { rooms: number; occupied: number }
  dormitories: {
    id: string
    name: string
    income: number
    expenses: number
    net: number
    billed: number
    outstanding: number
  }[]
}
