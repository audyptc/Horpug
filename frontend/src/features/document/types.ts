export type DocumentCategory = 'contract' | 'id_card' | 'receipt' | 'other'

export type ApiDocument = {
  id: string
  dormitory_id: string
  dormitory_name?: string
  tenant_id?: string
  tenant_name?: string
  room_id?: string
  room_number?: string
  name: string
  category: DocumentCategory
  file_url: string
  uploaded_date: string
  note: string
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}
