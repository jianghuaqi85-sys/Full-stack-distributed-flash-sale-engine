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

export type OrderStatus = 'pending' | 'confirmed' | 'paid' | 'shipped' | 'delivered' | 'cancelled'

export interface Order {
  id: number
  user_id?: number
  product_id: number
  quantity: number
  total: number
  status: OrderStatus
  created_at?: string
  updated_at?: string
}

export interface OrderListResponse {
  data: Order[]
  total: number
  page: number
  limit: number
}

export interface CreateOrderRequest {
  product_id: number
  quantity: number
}

export interface UpdateOrderRequest {
  status: OrderStatus
}
