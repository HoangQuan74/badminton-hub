import type { Tokens } from './auth.types'

const ACCESS_TOKEN = 'badminton_hub.access_token'
const REFRESH_TOKEN = 'badminton_hub.refresh_token'

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
