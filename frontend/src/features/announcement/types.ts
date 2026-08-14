export type ApiAnnouncement = {
  id: string
  dormitory_id: string
  dormitory_name?: string
  title: string
  content: string
  is_published: boolean
  published_date: string
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}
