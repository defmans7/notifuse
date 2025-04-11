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
  DeleteTemplateResponse,
  CompileTemplateRequest,
  CompileTemplateResponse
} from './types'

export const templatesApi = {
  list: async (params: GetTemplatesRequest): Promise<GetTemplatesResponse> => {
    const response = await api.get<GetTemplatesResponse>(
      `/api/templates.list?workspace_id=${params.workspace_id}`
    )
    return response
  },
  get: async (params: GetTemplateRequest): Promise<GetTemplateResponse> => {
    let url = `/api/templates.get?workspace_id=${params.workspace_id}&id=${params.id}`
    if (params.version) {
      url += `&version=${params.version}`
    }
    const response = await api.get<GetTemplateResponse>(url)
    return response
  },
  create: async (params: CreateTemplateRequest): Promise<CreateTemplateResponse> => {
    const response = await api.post<CreateTemplateResponse>(`/api/templates.create`, params)
    return response
  },
  update: async (params: UpdateTemplateRequest): Promise<UpdateTemplateResponse> => {
    const response = await api.post<UpdateTemplateResponse>(`/api/templates.update`, params)
    return response
  },
  delete: async (params: DeleteTemplateRequest): Promise<DeleteTemplateResponse> => {
    const response = await api.post<DeleteTemplateResponse>(`/api/templates.delete`, params)
    return response
  },
  compile: async (params: CompileTemplateRequest): Promise<CompileTemplateResponse> => {
    const response = await api.post<CompileTemplateResponse>(`/api/templates.compile`, params)
    return response
  }
}
