export type ApiActivityLog = {
  id: string
  user_id?: string
  username?: string
  action: string
  entity_type: string
  entity_id?: string
  description: string
  ip_address: string
  created_at: string
}
