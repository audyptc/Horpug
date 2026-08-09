export type ParcelStatus = 'pending' | 'picked_up'

export interface ApiParcel {
  id: string
  tracking_number: string
  recipient_name: string
  room_number: string
  status: ParcelStatus
  received_date: string
  picked_up_date: string | null
  note: string
  created_at: string
  updated_at: string
}

export interface CreateParcelPayload {
  tracking_number: string
  recipient_name: string
  room_number: string
  status: ParcelStatus
  received_date: string
  picked_up_date: string | null
  note: string
}

export interface UpdateParcelPayload {
  tracking_number: string
  recipient_name: string
  room_number: string
  status: ParcelStatus
  received_date: string
  picked_up_date: string | null
  note: string
}
