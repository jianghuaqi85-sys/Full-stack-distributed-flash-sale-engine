import request from './request'

export interface TicketTransfer {
  id: number
  ticket_id: number
  from_user_id: number
  to_user_id: number
  status: string
  transfer_type: string
  price: number
  reason: string
  created_at: string
}

export interface RequestTransferInput {
  ticket_id: number
  to_user_id: number
  reason?: string
}

export function requestTransfer(data: RequestTransferInput) {
  return request.post('/api/transfer', data)
}

export function directGift(data: RequestTransferInput) {
  return request.post('/api/transfer/gift', data)
}

export function getTransferHistory() {
  return request.get<{ data: TicketTransfer[] }>('/api/transfer/history')
}

export function getPendingTransfers() {
  return request.get<{ data: TicketTransfer[] }>('/admin/transfer/pending')
}

export function approveTransfer(id: number) {
  return request.post(`/admin/transfer/${id}/approve`)
}

export function rejectTransfer(id: number, reason?: string) {
  return request.post(`/admin/transfer/${id}/reject`, { reason })
}
