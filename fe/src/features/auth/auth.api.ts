import { http, publicHttp } from '@/lib/http'
import { portal } from '@/app/portal'
import { tokenStorage } from './auth.storage'
import type { AdminAuthResponse, AuthResponse, LoginPayload, MeResponse, RefreshResponse, RegisterPayload, User } from './auth.types'

const authPrefix = portal === 'admin' ? '/api/v1/admin/auth' : portal === 'tenant' ? '/api/v1/tenant/auth' : '/api/v1/auth'
const mePath = portal === 'admin' ? '/api/v1/admin/me' : portal === 'tenant' ? '/api/v1/tenant/me' : '/api/v1/me'

function staffAsUser(staff: AdminAuthResponse['staff']): User {
  return {
    id: staff.id,
    email: staff.email,
    full_name: staff.full_name,
    phone: '',
    status: staff.status,
    created_at: staff.created_at,
    roles: ['admin']
  }
}

export const authApi = {
  async login(payload: LoginPayload) {
    if (portal === 'admin') {
      const { data } = await publicHttp.post<AdminAuthResponse>(`${authPrefix}/login`, payload)
      tokenStorage.save(data.tokens)
      return { user: staffAsUser(data.staff), tokens: data.tokens }
    }
    const { data } = await publicHttp.post<AuthResponse>(`${authPrefix}/login`, payload)
    tokenStorage.save(data.tokens)
    return data
  },
  async register(payload: RegisterPayload) {
    const { data } = await publicHttp.post<AuthResponse>('/api/v1/auth/register', payload)
    tokenStorage.save(data.tokens)
    return data
  },
  async me() {
    if (portal === 'admin') {
      const { data } = await http.get<AdminAuthResponse['staff']>(mePath)
      return staffAsUser(data)
    }
    const { data } = await http.get<MeResponse>(mePath)
    return data.user
  },
  async logout() {
    const refreshToken = tokenStorage.refresh()
    if (refreshToken) await publicHttp.post<void>(`${authPrefix}/logout`, { refresh_token: refreshToken })
  },
  async refresh() {
    const { data } = await publicHttp.post<RefreshResponse>(`${authPrefix}/refresh`, {
      refresh_token: tokenStorage.refresh()
    })
    tokenStorage.save(data.tokens)
    return data.tokens
  }
}
