import { api } from './client'
import {
  CreateWorkspaceRequest,
  CreateWorkspaceResponse,
  ListWorkspacesResponse,
  GetWorkspaceResponse,
  UpdateWorkspaceRequest,
  UpdateWorkspaceResponse,
  DeleteWorkspaceRequest,
  DeleteWorkspaceResponse,
  GetWorkspaceMembersResponse
} from './types'

interface DetectFaviconResponse {
  iconUrl: string
  coverUrl?: string
}

export const workspaceService = {
  list: () => api.get<ListWorkspacesResponse>('/api/workspaces.list'),

  get: (id: string) => api.get<GetWorkspaceResponse>(`/api/workspaces.get?id=${id}`),

  create: (data: CreateWorkspaceRequest) =>
    api.post<CreateWorkspaceResponse>('/api/workspaces.create', data),

  update: (data: UpdateWorkspaceRequest) =>
    api.post<UpdateWorkspaceResponse>('/api/workspaces.update', data),

  delete: (data: DeleteWorkspaceRequest) =>
    api.post<DeleteWorkspaceResponse>('/api/workspaces.delete', data),

  detectFavicon: (url: string) => api.post<DetectFaviconResponse>('/api/detect-favicon', { url }),

  getMembers: (id: string) =>
    api.get<GetWorkspaceMembersResponse>(`/api/workspaces.members?id=${id}`)
}
