import { request } from '@/utils/request'

export interface LoginParams {
  username: string
  password: string
}

export interface RegisterParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  username: string
  expire_at: string
  default_pwd: boolean
  project_cache_ttl_seconds: number
  session_idle_ttl_seconds: number
}

export interface UserInfo {
  id: number
  username: string
  project_cache_ttl_seconds: number
  session_idle_ttl_seconds: number
}

export const authApi = {
  async login(data: LoginParams): Promise<LoginResult> {
    return request({
      url: '/auth/login',
      method: 'post',
      data
    })
  },

  async register(data: RegisterParams): Promise<{ message: string }> {
    return request({
      url: '/auth/register',
      method: 'post',
      data
    })
  },

  logout(): Promise<{ ok: boolean }> {
    return request({
      url: '/auth/logout',
      method: 'post'
    })
  },

  getMe(): Promise<UserInfo> {
    return request({
      url: '/auth/me',
      method: 'get'
    })
  },

  changePassword(data: { old_password: string; new_password: string }): Promise<{ message: string }> {
    return request({
      url: '/auth/change-password',
      method: 'post',
      data
    })
  }
}
