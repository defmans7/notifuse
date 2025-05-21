import { api } from './client'
import {
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

// Define the API interfaces
export interface TemplatesApi {
  list: (params: GetTemplatesRequest) => Promise<GetTemplatesResponse>
  get: (params: GetTemplateRequest) => Promise<GetTemplateResponse>
  create: (params: CreateTemplateRequest) => Promise<CreateTemplateResponse>
  update: (params: UpdateTemplateRequest) => Promise<UpdateTemplateResponse>
  delete: (params: DeleteTemplateRequest) => Promise<DeleteTemplateResponse>
  compile: (params: CompileTemplateRequest) => Promise<CompileTemplateResponse>
}

export const templatesApi: TemplatesApi = {
  list: async (params: GetTemplatesRequest): Promise<GetTemplatesResponse> => {
    let url = `/api/templates.list?workspace_id=${params.workspace_id}`
    if (params.category) {
      url += `&category=${params.category}`
    }
    const response = await api.get<GetTemplatesResponse>(url)
    return response
  },
  get: async (params: GetTemplateRequest): Promise<GetTemplateResponse> => {
    let url = `/api/templates.get?workspace_id=${params.workspace_id}&id=${params.id}&version=${params.version || 0}`
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
