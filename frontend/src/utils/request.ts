import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { useAuthStore } from '@/stores/auth'

const service: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000
})

service.interceptors.request.use(
  (config) => {
    const auth = useAuthStore()
    if (auth.token) {
      config.headers['Authorization'] = `Bearer ${auth.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

let isLoggingOut = false

service.interceptors.response.use(
  (response: AxiosResponse) => {
    if (response.config.responseType === 'blob') {
      return response.data
    }
    return response.data
  },
  (error) => {
    const status = error.response?.status
    const data = error.response?.data

    if (status === 401 && !isLoggingOut) {
      isLoggingOut = true
      window.$message?.error('当前用户登录已过期，请重新登录')
      const auth = useAuthStore()
      auth.clearSession()
      isLoggingOut = false
      return Promise.reject(new Error('登录已过期'))
    }

    const message = data?.error || data?.message || error.message || '网络错误'
    window.$message?.error(message)
    return Promise.reject(error)
  }
)

export function request<T = any>(config: AxiosRequestConfig): Promise<T> {
  return service(config) as Promise<T>
}

export default service
