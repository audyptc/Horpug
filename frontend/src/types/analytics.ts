export interface MonthlyRevenue {
  month: string
  revenue: number
}

export interface MonthlyCount {
  month: string
  count: number
}

export interface BillStatusCount {
  unpaid: number
  paid: number
  overdue: number
}

export interface ApiAnalyticsSummary {
  monthly_revenue: MonthlyRevenue[]
  monthly_tenants: MonthlyCount[]
  bill_status: BillStatusCount
  total_revenue: number
  avg_monthly: number
}

export interface ApiDashboardSummary {
  total_rooms: number
  occupied_rooms: number
  available_rooms: number
  maintenance_rooms: number
  occupancy_rate: number
  total_tenants: number
  active_contracts: number
  expiring_contracts: number
  unpaid_bills: number
  overdue_bills: number
  unpaid_amount: number
  overdue_amount: number
  revenue_this_month: number
}
