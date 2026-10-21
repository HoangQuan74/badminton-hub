export type UserStatus = 'active' | 'suspended'
export type UserRole = 'customer' | 'tenant_owner' | 'tenant_staff' | 'admin'

export interface User {
  id: string
  email: string
  full_name: string
  phone: string
  status: UserStatus
  created_at: string
  /** Optional until the User Gateway exposes roles in its public user response. */
  roles?: UserRole[]
}

export interface Tokens {
  access_token: string
  refresh_token: string
  token_type: string
  access_expires_at: string
  refresh_expires_at: string
}

export interface AuthResponse { user: User; tokens: Tokens }
export interface AdminAuthResponse { staff: Staff; tokens: Tokens }
export interface RefreshResponse { tokens: Tokens }
export interface MeResponse { user: User }

export interface Staff {
  id: string
  email: string
  full_name: string
  is_admin: boolean
  role_id: string
  role_name: string
  status: UserStatus
  created_at: string
}

export interface LoginPayload { email: string; password: string }
export interface RegisterPayload extends LoginPayload {
  full_name: string
  phone?: string
}

export interface ApiErrorBody {
  error?: { code?: string; message?: string }
}
