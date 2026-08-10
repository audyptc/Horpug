export interface ApiResponse<T> {
  success: boolean
  data: T
}

export interface ApiErrorResponse {
  success: false
  message: string
}

export interface ApiPaginatedResponse<T> {
  success: boolean
  data: T[]
  meta: PaginationMeta
}

export interface PaginationMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}
