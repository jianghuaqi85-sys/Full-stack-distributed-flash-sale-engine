import request from './request'
import type { User } from '../types'

export const getProfile = () =>
  request.get<User>('/api/profile')

export const changePassword = (data: { old_password: string; new_password: string }) =>
  request.post<{ message: string }>('/api/profile/password', data)
