import { api } from './client'

export interface ListContactsRequest {
  workspace_id: string
  // Optional filters
  email?: string
  external_id?: string
  first_name?: string
  last_name?: string
  phone?: string
  country?: string
  language?: string
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
  contact_lists: {
    email: string
    list_id: string
    status: string
    created_at: string
    updated_at: string
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
    searchParams.append('workspace_id', params.workspace_id)

    // Add optional params if they exist
    if (params.email) searchParams.append('email', params.email)
    if (params.external_id) searchParams.append('external_id', params.external_id)
    if (params.first_name) searchParams.append('first_name', params.first_name)
    if (params.last_name) searchParams.append('last_name', params.last_name)
    if (params.phone) searchParams.append('phone', params.phone)
    if (params.country) searchParams.append('country', params.country)
    if (params.language) searchParams.append('language', params.language)
    if (params.limit) searchParams.append('limit', params.limit.toString())
    if (params.cursor) searchParams.append('cursor', params.cursor)

    return api.get<ListContactsResponse>(`/api/contacts.list?${searchParams.toString()}`)
  }
}
