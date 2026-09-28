import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { admin, clearToken, getToken, setToken } from '@/lib/api'

export const useAdminStore = defineStore('admin', () => {
  const token = ref<string | null>(getToken())

  const isAuthenticated = computed(() => !!token.value)

  async function login(username: string, password: string) {
    const { token: t } = await admin.login(username, password)
    setToken(t)
    token.value = t
  }

  function logout() {
    clearToken()
    token.value = null
  }

  // 401 em qualquer chamada derruba a sessão (interceptor dispara o evento)
  window.addEventListener('brick:unauthorized', () => {
    token.value = null
  })

  return { token, isAuthenticated, login, logout }
})
