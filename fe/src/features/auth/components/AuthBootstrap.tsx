import { useEffect, useState, type ReactNode } from 'react'
import { authApi } from '../auth.api'
import { tokenStorage } from '../auth.storage'
import { useAuthStore } from '../auth.store'

export function AuthBootstrap({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false)
  const setUser = useAuthStore((state) => state.setUser)
  const setInitialized = useAuthStore((state) => state.setInitialized)

  useEffect(() => {
    let active = true
    async function bootstrap() {
      if (!tokenStorage.access() && !tokenStorage.refresh()) {
        setInitialized(true)
        setReady(true)
        return
      }
      try {
        const user = await authApi.me()
        if (active) setUser(user)
      } catch {
        tokenStorage.clear()
        if (active) setUser(null)
      } finally {
        if (active) {
          setInitialized(true)
          setReady(true)
        }
      }
    }
    void bootstrap()
    return () => { active = false }
  }, [setInitialized, setUser])

  if (!ready) {
    return (
      <div className="app-loader" role="status" aria-live="polite">
        <span className="brand-mark" aria-hidden="true">B</span>
        <span>Đang chuẩn bị sân cho bạn…</span>
      </div>
    )
  }
  return children
}
