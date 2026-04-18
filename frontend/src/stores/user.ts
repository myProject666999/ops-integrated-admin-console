import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type LoginParams, type UserInfo } from '@/api/auth'
import router from '@/router'
import { useAuthStore } from './auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(null)
  const user = ref<UserInfo | null>(null)
  
  const isLogin = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '')
  
  async function login(params: LoginParams) {
    const res = await authApi.login(params)
    token.value = res.token
    user.value = {
      id: 0,
      username: res.username,
      project_cache_ttl_seconds: res.project_cache_ttl_seconds,
      session_idle_ttl_seconds: res.session_idle_ttl_seconds
    }
    
    const auth = useAuthStore()
    auth.setSession(
      res.token,
      res.username,
      res.expire_at,
      res.session_idle_ttl_seconds
    )
    
    return res
  }
  
  async function logout() {
    const hadToken = !!token.value
    token.value = null
    user.value = null
    
    if (hadToken) {
      try {
        await authApi.logout()
      } catch (error) {
        // ignore
      }
    }
    
    const auth = useAuthStore()
    auth.clearSession()
    router.push('/login')
  }
  
  async function getMe() {
    const res = await authApi.getMe()
    user.value = res
    return res
  }
  
  return {
    token,
    user,
    isLogin,
    username,
    login,
    logout,
    getMe
  }
}, {
  persist: {
    key: 'ops-user',
    paths: ['token']
  }
})
