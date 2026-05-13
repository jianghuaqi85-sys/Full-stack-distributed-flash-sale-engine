import request from './request'
import type { User } from '../types'

export const getProfile = () =>
  request.get<User>('/api/profile')
