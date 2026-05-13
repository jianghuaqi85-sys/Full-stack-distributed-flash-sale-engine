import request from './request'

export interface QueuePosition {
  position: number
  total_ahead: number
  estimated_wait: number
  status: string
}

export function joinQueue(eventId: number) {
  return request.post<QueuePosition>(`/api/queue/${eventId}/join`)
}

export function getQueuePosition(eventId: number) {
  return request.get<QueuePosition>(`/api/queue/${eventId}/position`)
}

export function leaveQueue(eventId: number) {
  return request.post(`/api/queue/${eventId}/leave`)
}
