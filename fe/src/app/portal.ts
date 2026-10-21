export type Portal = 'user' | 'tenant' | 'admin'

export const portal = (import.meta.env.VITE_PORTAL || 'user') as Portal

export const portalUrls: Record<Portal, string> = {
  user: import.meta.env.VITE_USER_URL || 'http://localhost:5173',
  tenant: import.meta.env.VITE_TENANT_URL || 'http://localhost:5174',
  admin: import.meta.env.VITE_ADMIN_URL || 'http://localhost:5175'
}

export function isPortal(value: string): value is Portal {
  return value === 'user' || value === 'tenant' || value === 'admin'
}
