import request from './request'
import type { LoginResponse, RegisterRequest } from '../types'

export const login = (data: { username: string; password: string }) =>
  request.post<LoginResponse>('/api/auth/login', data)

export const register = (data: RegisterRequest) =>
  request.post<{ message: string }>('/api/auth/register', data)
