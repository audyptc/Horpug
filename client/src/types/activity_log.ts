export type ActivityAction = 'create' | 'update' | 'delete'

export interface ApiActivityLog {
  id: string
  actor_id: string | null
  actor_name: string
  action: ActivityAction
  entity_type: string
  entity_id: string
  new_value: unknown
  created_at: string
}
