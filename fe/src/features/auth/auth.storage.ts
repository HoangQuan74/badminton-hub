import type { Tokens } from './auth.types'
import { portal } from '@/app/portal'

const ACCESS_TOKEN = `badminton_hub.${portal}.access_token`
const REFRESH_TOKEN = `badminton_hub.${portal}.refresh_token`

export const tokenStorage = {
  access: () => localStorage.getItem(ACCESS_TOKEN),
  refresh: () => localStorage.getItem(REFRESH_TOKEN),
  save: (tokens: Tokens) => {
    localStorage.setItem(ACCESS_TOKEN, tokens.access_token)
    localStorage.setItem(REFRESH_TOKEN, tokens.refresh_token)
  },
  clear: () => {
    localStorage.removeItem(ACCESS_TOKEN)
    localStorage.removeItem(REFRESH_TOKEN)
  }
}
