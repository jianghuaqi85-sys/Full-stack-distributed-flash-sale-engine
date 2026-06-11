export interface User {
  id: number
  username: string
  email: string
  role: string
}

export interface LoginResponse {
  access_token: string
  token_type: string
  expires_in: number
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
}

// 票务状态
export type TicketStatus = 'reserved' | 'paid' | 'used' | 'expired' | 'cancelled'

// 活动状态
export type EventStatus = 'draft' | 'on_sale' | 'off_sale' | 'ended'

// 票务
export interface Ticket {
  id: number
  order_no: string
  event_id: number
  ticket_name: string
  quantity: number
  total_price: number
  status: TicketStatus
  qr_code?: string
  created_at: string
}

// 活动
export interface Event {
  id: number
  title: string
  description: string
  location: string
  cover_image: string
  start_time: string
  end_time: string
  status: EventStatus
  total_stock: number
  ticket_types?: TicketType[]
}

// 票种
export interface TicketType {
  id: number
  name: string
  price: number
  stock: number
  max_per_user: number
  sort_order: number
}

// 二手市场挂单
export interface MarketplaceListing {
  id: number
  ticket_id: number
  event_id: number
  ticket_name: string
  event_title?: string
  seller_id: number
  seller_name?: string
  price: number
  status: 'active' | 'sold' | 'cancelled'
  buyer_id?: number
  description: string
  created_at: string
}

// 分页响应
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
}
