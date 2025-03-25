import { api } from './client'

export interface ListContactsRequest {
  workspaceId: string
  // Optional filters
  email?: string
  externalId?: string
  firstName?: string
  lastName?: string
  phone?: string
  country?: string
  // Pagination
  limit?: number
  cursor?: string
}

export interface Contact {
  id: string
  email: string
  first_name: string
  last_name: string
  country_code: string
  subscriptions: {
    id: string
    name: string
  }[]
  created_at: string
  updated_at: string
}

export interface ListContactsResponse {
  contacts: Contact[]
  next_cursor?: string
}

export const contactsApi = {
  list: async (params: ListContactsRequest): Promise<ListContactsResponse> => {
    const searchParams = new URLSearchParams()

    // Add required param
    searchParams.append('workspaceId', params.workspaceId)

    // Add optional params if they exist
    if (params.email) searchParams.append('email', params.email)
    if (params.externalId) searchParams.append('externalId', params.externalId)
    if (params.firstName) searchParams.append('firstName', params.firstName)
    if (params.lastName) searchParams.append('lastName', params.lastName)
    if (params.phone) searchParams.append('phone', params.phone)
    if (params.country) searchParams.append('country', params.country)
    if (params.limit) searchParams.append('limit', params.limit.toString())
    if (params.cursor) searchParams.append('cursor', params.cursor)

    return api.get<ListContactsResponse>(`/api/contacts.list?${searchParams.toString()}`)
  }
}
