import request from './request'

export interface Show {
  id: number
  event_id: number
  name: string
  show_time: string
  end_time: string
  status: string
  stock: number
  sold_count: number
  sort_order: number
}

export function getEventShows(eventId: number) {
  return request.get<{ data: Show[] }>(`/api/events/${eventId}/shows`)
}

export function getShow(id: number) {
  return request.get<Show>(`/api/shows/${id}`)
}

export function createShow(eventId: number, data: {
  name: string
  show_time: string
  end_time: string
  stock: number
  sort_order?: number
}) {
  return request.post(`/admin/events/${eventId}/shows`, data)
}

export function updateShow(id: number, data: {
  name: string
  show_time: string
  end_time: string
  stock: number
  sort_order?: number
}) {
  return request.put(`/admin/events/shows/${id}`, data)
}

export function deleteShow(id: number) {
  return request.delete(`/admin/events/shows/${id}`)
}

export function publishShow(id: number) {
  return request.post(`/admin/events/shows/${id}/publish`)
}

export function unpublishShow(id: number) {
  return request.post(`/admin/events/shows/${id}/unpublish`)
}
