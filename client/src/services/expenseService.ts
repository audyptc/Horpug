import api from '@/lib/axios'
import type {
  ApiPaginatedResponse,
  ApiResponse,
  ApiExpense,
  CreateExpensePayload,
  UpdateExpensePayload,
} from '@/types/api'

export const expenseService = {
  async list(page = 1, perPage = 20): Promise<ApiPaginatedResponse<ApiExpense>> {
    const offset = (page - 1) * perPage
    const { data } = await api.get<ApiPaginatedResponse<ApiExpense>>(
      `/expenses/?limit=${perPage}&offset=${offset}`
    )
    return data
  },

  async getById(id: string): Promise<ApiExpense> {
    const { data } = await api.get<ApiResponse<ApiExpense>>(`/expenses/${id}`)
    return data.data
  },

  async create(payload: CreateExpensePayload): Promise<ApiExpense> {
    const { data } = await api.post<ApiResponse<ApiExpense>>('/expenses/', payload)
    return data.data
  },

  async update(id: string, payload: UpdateExpensePayload): Promise<ApiExpense> {
    const { data } = await api.put<ApiResponse<ApiExpense>>(`/expenses/${id}`, payload)
    return data.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/expenses/${id}`)
  },
}
