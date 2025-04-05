import { api } from './client'
import type {
  CreateListRequest,
  GetListsRequest,
  GetListRequest,
  UpdateListRequest,
  DeleteListRequest,
  GetListsResponse,
  GetListResponse,
  CreateListResponse,
  UpdateListResponse,
  DeleteListResponse
} from './types'

export const listsApi = {
  create: async (params: CreateListRequest): Promise<CreateListResponse> => {
    return api.post('/api/lists.create', params)
  },

  list: async (params: GetListsRequest): Promise<GetListsResponse> => {
    const searchParams = new URLSearchParams()

    // Add required param
    searchParams.append('workspace_id', params.workspace_id)

    return api.get<GetListsResponse>(`/api/lists.list?${searchParams.toString()}`)
  },

  get: async (params: GetListRequest): Promise<GetListResponse> => {
    const searchParams = new URLSearchParams()

    // Add required params
    searchParams.append('workspace_id', params.workspace_id)
    searchParams.append('id', params.id)

    return api.get<GetListResponse>(`/api/lists.get?${searchParams.toString()}`)
  },

  update: async (params: UpdateListRequest): Promise<UpdateListResponse> => {
    return api.post('/api/lists.update', params)
  },

  delete: async (params: DeleteListRequest): Promise<DeleteListResponse> => {
    return api.post('/api/lists.delete', params)
  }
}
