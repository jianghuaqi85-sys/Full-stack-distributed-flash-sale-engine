import request from './request'

export interface Ticket {
  id: number
  order_no: string
  event_id: number
  ticket_name: string
  quantity: number
  total_price: number
  status: string
  qr_code?: string
  created_at: string
}

export interface TicketListResponse {
  data: Ticket[]
  total: number
  page: number
  limit: number
}

export function purchaseTicket(data: {
  event_id: number
  show_id?: number
  ticket_type_id: number
  quantity: number
}) {
  return request.post('/api/tickets/purchase', data)
}

export function getMyTickets(page = 1, limit = 10) {
  return request.get<TicketListResponse>('/api/tickets', { params: { page, limit } })
}

export function getTicketDetail(id: number) {
  return request.get<Ticket>(`/api/tickets/${id}`)
}

export function payTicket(id: number) {
  return request.post(`/api/tickets/${id}/pay`)
}

export function cancelTicket(id: number) {
  return request.post(`/api/tickets/${id}/cancel`)
}

export function useTicket(id: number) {
  return request.post(`/api/tickets/${id}/use`)
}
