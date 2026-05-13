import request from './request'
import type { Show } from './shows'

export interface TicketType {
  id: number
  name: string
  price: number
  stock: number
  max_per_user: number
  sort_order: number
}

export interface Event {
  id: number
  title: string
  description: string
  location: string
  cover_image: string
  start_time: string
  end_time: string
  status: string
  total_stock: number
  ticket_types?: TicketType[]
  shows?: Show[]
}

export type { Show }

export interface EventListResponse {
  data: Event[]
  total: number
  page: number
  limit: number
}

export function getEvents(page = 1, limit = 20) {
  return request.get<EventListResponse>('/api/events', { params: { page, limit } })
}

export function getEvent(id: number) {
  return request.get<Event>(`/api/events/${id}`)
}

export function getEventStock(id: number) {
  return request.get<Record<string, number>>(`/api/events/${id}/stock`)
}

export function createEvent(data: {
  title: string
  description?: string
  location: string
  cover_image?: string
  start_time: string
  end_time: string
}) {
  return request.post('/admin/events', data)
}

export function updateEvent(id: number, data: {
  title: string
  description?: string
  location: string
  cover_image?: string
  start_time: string
  end_time: string
}) {
  return request.put(`/admin/events/${id}`, data)
}

export function publishEvent(id: number) {
  return request.post(`/admin/events/${id}/publish`)
}

export function unpublishEvent(id: number) {
  return request.post(`/admin/events/${id}/unpublish`)
}

export function endEvent(id: number) {
  return request.post(`/admin/events/${id}/end`)
}

export function createTicketType(eventId: number, data: {
  name: string
  price: number
  stock: number
  max_per_user?: number
  sort_order?: number
}) {
  return request.post(`/admin/events/${eventId}/ticket-types`, data)
}

export function updateTicketType(id: number, data: {
  name: string
  price: number
  stock: number
  max_per_user?: number
  sort_order?: number
}) {
  return request.put(`/admin/events/ticket-types/${id}`, data)
}

export function deleteTicketType(id: number) {
  return request.delete(`/admin/events/ticket-types/${id}`)
}
