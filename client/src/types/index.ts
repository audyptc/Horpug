export interface Activity {
  id: string
  user: string
  action: string
  target: string
  time: string
  type: 'create' | 'update' | 'delete' | 'login'
}
