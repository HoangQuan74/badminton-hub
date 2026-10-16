import { http, publicHttp } from '@/lib/http'
import { tokenStorage } from './auth.storage'
import type { AuthResponse, LoginPayload, MeResponse, RefreshResponse, RegisterPayload } from './auth.types'

export const authApi = {
  async login(payload: LoginPayload) {
    const { data } = await publicHttp.post<AuthResponse>('/api/v1/auth/login', payload)
    tokenStorage.save(data.tokens)
    return data
  },
  async register(payload: RegisterPayload) {
    const { data } = await publicHttp.post<AuthResponse>('/api/v1/auth/register', payload)
    tokenStorage.save(data.tokens)
    return data
  },
  async me() {
    const { data } = await http.get<MeResponse>('/api/v1/me')
    return data.user
  },
  async logout() {
    const refreshToken = tokenStorage.refresh()
    if (refreshToken) await publicHttp.post<void>('/api/v1/auth/logout', { refresh_token: refreshToken })
  },
  async refresh() {
    const { data } = await publicHttp.post<RefreshResponse>('/api/v1/auth/refresh', {
      refresh_token: tokenStorage.refresh()
    })
    tokenStorage.save(data.tokens)
    return data.tokens
  }
}
