import request from './request'

export interface DashboardStats {
  total_events: number
  active_events: number
  total_tickets: number
  sold_tickets: number
  reserved_tickets: number
  total_revenue: number
  today_sales: number
  today_revenue: number
}

export interface SalesTrend {
  date: string
  count: number
  revenue: number
}

export interface TicketTypeStats {
  ticket_type_id: number
  ticket_type_name: string
  event_title: string
  sold_count: number
  revenue: number
}

export interface ConversionFunnel {
  page_views: number
  add_to_cart: number
  reserved: number
  paid: number
  used: number
}

export function getDashboardStats() {
  return request.get<DashboardStats>('/admin/stats/dashboard')
}

export function getSalesTrend(days: number = 7) {
  return request.get<{ data: SalesTrend[] }>('/admin/stats/sales-trend', { params: { days } })
}

export function getTicketTypeStats() {
  return request.get<{ data: TicketTypeStats[] }>('/admin/stats/ticket-types')
}

export function getConversionFunnel(eventId: number) {
  return request.get<ConversionFunnel>(`/admin/stats/funnel/${eventId}`)
}
