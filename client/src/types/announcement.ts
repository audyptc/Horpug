export type AnnouncementType = 'general' | 'maintenance' | 'payment' | 'emergency'

export interface ApiAnnouncement {
  id: string
  title: string
  content: string
  type: AnnouncementType
  is_pinned: boolean
  published_at: string
  expired_at: string | null
  created_at: string
  updated_at: string
}

export interface CreateAnnouncementPayload {
  title: string
  content: string
  type: AnnouncementType
  is_pinned: boolean
  published_at: string
  expired_at: string | null
}

export interface UpdateAnnouncementPayload {
  title: string
  content: string
  type: AnnouncementType
  is_pinned: boolean
  published_at: string
  expired_at: string | null
}
