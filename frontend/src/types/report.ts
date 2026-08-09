export interface IncomeReportRow {
  month: string
  rent_amount: number
  electric_amount: number
  water_amount: number
  other_amount: number
  total_amount: number
  bill_count: number
}

export interface ApiIncomeReport {
  rows: IncomeReportRow[]
  total_rent: number
  total_electric: number
  total_water: number
  total_other: number
  grand_total: number
}

export interface ExpenseReportRow {
  month: string
  category: string
  total_amount: number
  count: number
}

export interface ApiExpenseReport {
  rows: ExpenseReportRow[]
  grand_total: number
}

export interface OccupancyRow {
  room_number: string
  floor: number
  type: string
  status: string
  tenant_name: string
  rent_price: number
}

export interface ApiOccupancyReport {
  rows: OccupancyRow[]
  total: number
  occupied: number
  available: number
  maintenance: number
  occupancy_rate: number
}
