import axios from 'axios'
import { message } from 'antd'

const request = axios.create({
  baseURL: '',
  timeout: 10000,
})

request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

request.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    const errorMsg = error.response?.data?.error || '请求失败'
    message.error(errorMsg)
    return Promise.reject(error)
  }
)

export default request
