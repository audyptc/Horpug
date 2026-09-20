// Mirrors the categories the announcements endpoint accepts.
export type AnnouncementCategory = 'general' | 'urgent' | 'billing' | 'maintenance' | 'event'

export type ApiAnnouncement = {
  id: string
  dormitory_id: string
  dormitory_name?: string
  title: string
  content: string
  category: AnnouncementCategory
  // Pinned announcements are listed first by default.
  is_pinned: boolean
  is_published: boolean
  published_date: string
  // Whether the signed-in user has opened it; per user, not per announcement.
  is_read: boolean
  created_by?: string
  updated_by?: string
  created_at: string
  updated_at: string
}

export type ApiAnnouncementSummary = {
  unread_count: number
  // Whether the role may create, update or delete announcements.
  can_manage: boolean
}
