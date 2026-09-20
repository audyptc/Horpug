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
  // A document is either a link to an externally hosted file (file_url) or a
  // file uploaded to the server (has_file, fetched through the file endpoint);
  // file_url is empty for uploaded ones.
  file_url: string
  has_file: boolean
  file_name?: string
  file_mime?: string
  file_size?: number
  uploaded_date: string
  note: string
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}
