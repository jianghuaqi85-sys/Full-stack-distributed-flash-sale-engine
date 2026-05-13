import request from './request'

export interface MarketplaceListing {
  id: number
  ticket_id: number
  event_id: number
  ticket_name: string
  event_title: string
  seller_id: number
  seller_name: string
  price: number
  status: string
  buyer_id: number
  description: string
  created_at: string
}

export function getActiveListings(page = 1, limit = 20) {
  return request.get<{ data: MarketplaceListing[]; total: number }>('/api/marketplace', { params: { page, limit } })
}

export function getListing(id: number) {
  return request.get<MarketplaceListing>(`/api/marketplace/${id}`)
}

export function getEventListings(eventId: number, page = 1, limit = 20) {
  return request.get<{ data: MarketplaceListing[]; total: number }>(`/api/marketplace/event/${eventId}`, { params: { page, limit } })
}

export function getMyListings() {
  return request.get<{ data: MarketplaceListing[] }>('/api/marketplace/my')
}

export function getMyPurchases() {
  return request.get<{ data: MarketplaceListing[] }>('/api/marketplace/purchases')
}

export function createListing(data: {
  ticket_id: number
  price: number
  description?: string
}) {
  return request.post('/api/marketplace', data)
}

export function buyListing(id: number) {
  return request.post(`/api/marketplace/${id}/buy`)
}

export function cancelListing(id: number) {
  return request.post(`/api/marketplace/${id}/cancel`)
}
