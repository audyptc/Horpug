export type DocumentCategory = 'contract' | 'id_card' | 'house_registration' | 'receipt' | 'other'

export interface ApiDocument {
  id: string
  title: string
  category: DocumentCategory
  tenant_id: string | null
  file_url: string
  issue_date: string | null
  expiry_date: string | null
  note: string
  created_at: string
  updated_at: string
  tenant_first_name: string
  tenant_last_name: string
}

export interface CreateDocumentPayload {
  title: string
  category: DocumentCategory
  tenant_id: string | null
  file_url: string
  issue_date: string | null
  expiry_date: string | null
  note: string
}

export interface UpdateDocumentPayload {
  title: string
  category: DocumentCategory
  tenant_id: string | null
  file_url: string
  issue_date: string | null
  expiry_date: string | null
  note: string
}
