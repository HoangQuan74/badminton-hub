import { http } from '@/lib/http'
import { catalogApi, type Venue } from '@/features/catalog/catalog.api'

export interface VenuePayload { name: string; slug: string; address: string; timezone: string }

export const venueApi = {
  list: catalogApi.listTenantVenues,
  async create(organizationId: string, payload: VenuePayload) {
    const { data } = await http.post<{ venue: Omit<Venue, 'courts'> }>(`/api/v1/tenant/${organizationId}/venues`, payload)
    return data.venue
  },
  async update(id: string, payload: VenuePayload) {
    const { data } = await http.patch<{ venue: Omit<Venue, 'courts'> }>(`/api/v1/tenant/venues/${id}`, payload)
    return data.venue
  },
  async archive(id: string) { await http.delete(`/api/v1/tenant/venues/${id}`) }
}
