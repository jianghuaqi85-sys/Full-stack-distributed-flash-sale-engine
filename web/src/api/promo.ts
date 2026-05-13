import request from './request'

export interface PromoCode {
  id: number
  code: string
  event_id: number
  discount_type: string
  discount_value: number
  min_amount: number
  max_uses: number
  used_count: number
  start_time: string
  end_time: string
  is_active: boolean
}

export interface ValidatePromoResponse {
  code: string
  discount_type: string
  discount_value: number
  discount: number
  final_amount: number
}

export function validatePromoCode(code: string, amount: number) {
  return request.post<ValidatePromoResponse>('/api/promo/validate', { code, amount })
}

export function getPromoCodes(eventId: number) {
  return request.get<{ data: PromoCode[] }>(`/api/promo/${eventId}`)
}

export function createPromoCode(data: {
  code: string
  event_id?: number
  discount_type: 'percent' | 'fixed'
  discount_value: number
  min_amount?: number
  max_uses?: number
  start_time?: string
  end_time?: string
}) {
  return request.post('/admin/promo', data)
}

export function deletePromoCode(id: number) {
  return request.delete(`/admin/promo/${id}`)
}
