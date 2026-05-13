import request from './request'

export interface WaitlistPosition {
  position: number
  status: string
  message?: string
}

export function joinWaitlist(eventId: number) {
  return request.post<WaitlistPosition>(`/api/waitlist/${eventId}/join`)
}

export function getWaitlistPosition(eventId: number) {
  return request.get<WaitlistPosition>(`/api/waitlist/${eventId}/position`)
}

export function leaveWaitlist(eventId: number) {
  return request.post(`/api/waitlist/${eventId}/leave`)
}
