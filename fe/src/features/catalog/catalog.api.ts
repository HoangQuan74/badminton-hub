import type { CourtStatus as CourtStatusType, VenueStatus } from '@/constants/domain'
import { http, publicHttp } from '@/lib/http'

export type CourtStatus = CourtStatusType

export interface Court {
  id: string
  venue_id: string
  organization_id: string
  name: string
  status: CourtStatus
  slot_interval_minutes: number
  created_at: string
  updated_at: string
}

export interface Venue {
  id: string
  organization_id: string
  name: string
  slug: string
  address: string
  timezone: string
  status: VenueStatus
  courts: Court[]
}

interface ListResponse<T> { items?: T[] }
export interface CourtPayload { name: string; slot_interval_minutes: number; status?: CourtStatus }

function mapVenueList(response: ListResponse<Venue>): Venue[] {
  return (response.items ?? []).map((venue) => ({ ...venue, courts: venue.courts ?? [] }))
}

export const catalogApi = {
  async listPublicVenues() {
    const { data } = await publicHttp.get<ListResponse<Venue>>('/api/v1/venues')
    return mapVenueList(data)
  },
  async listTenantVenues() {
    const { data } = await http.get<ListResponse<Venue>>('/api/v1/tenant/venues')
    return mapVenueList(data)
  },
  async listTenantCourts(venueId: string) {
    const { data } = await http.get<ListResponse<Court>>(`/api/v1/tenant/venues/${venueId}/courts`)
    return data.items ?? []
  },
  async createCourt(venueId: string, payload: CourtPayload) {
    const { data } = await http.post<{ court: Court }>(`/api/v1/tenant/venues/${venueId}/courts`, payload)
    return data.court
  },
  async updateCourt(courtId: string, payload: CourtPayload) {
    const { data } = await http.patch<{ court: Court }>(`/api/v1/tenant/courts/${courtId}`, payload)
    return data.court
  },
  async archiveCourt(courtId: string) {
    await http.delete(`/api/v1/tenant/courts/${courtId}`)
  }
}
