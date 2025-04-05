import { api } from './client'
import type {
  GetTemplatesRequest,
  GetTemplatesResponse,
  GetTemplateRequest,
  GetTemplateResponse,
  CreateTemplateRequest,
  CreateTemplateResponse,
  UpdateTemplateRequest,
  UpdateTemplateResponse,
  DeleteTemplateRequest,
  DeleteTemplateResponse
} from './types'

export const templatesApi = {
  list: async (params: GetTemplatesRequest): Promise<GetTemplatesResponse> => {
    const response = await api.get<GetTemplatesResponse>(
      `/workspaces/${params.workspace_id}/templates`
    )
    return response
  },
  get: async (params: GetTemplateRequest): Promise<GetTemplateResponse> => {
    const response = await api.get<GetTemplateResponse>(
      `/workspaces/${params.workspace_id}/templates/${params.id}`
    )
    return response
  },
  create: async (params: CreateTemplateRequest): Promise<CreateTemplateResponse> => {
    const response = await api.post<CreateTemplateResponse>(
      `/workspaces/${params.workspace_id}/templates`,
      params
    )
    return response
  },
  update: async (params: UpdateTemplateRequest): Promise<UpdateTemplateResponse> => {
    const response = await api.put<UpdateTemplateResponse>(
      `/workspaces/${params.workspace_id}/templates/${params.id}`,
      params
    )
    return response
  },
  delete: async (params: DeleteTemplateRequest): Promise<DeleteTemplateResponse> => {
    const response = await api.delete<DeleteTemplateResponse>(
      `/workspaces/${params.workspace_id}/templates/${params.id}`
    )
    return response
  }
}
