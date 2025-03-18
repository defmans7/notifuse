// Authentication types
export interface SignInRequest {
  email: string
}

export interface SignInResponse {
  message: string
  code?: string
}

export interface VerifyCodeRequest {
  email: string
  code: string
}

export interface VerifyResponse {
  token: string
}

export interface GetCurrentUserResponse {
  user: {
    id: string
    email: string
    timezone: string
  }
  workspaces: Workspace[]
}

// Workspace types
export interface WorkspaceSettings {
  name: string
  url: string
  logo_url: string | null
  timezone: string
}

export interface CreateWorkspaceRequest {
  id: string
  settings: WorkspaceSettings
}

export interface Workspace {
  id: string
  settings: WorkspaceSettings
  created_at: string
  updated_at: string
}

export interface CreateWorkspaceResponse {
  workspace: Workspace
}

export interface ListWorkspacesResponse {
  workspaces: Workspace[]
}

export interface GetWorkspaceResponse {
  workspace: Workspace
}

export interface UpdateWorkspaceRequest {
  id: string
  settings: Partial<WorkspaceSettings>
}

export interface UpdateWorkspaceResponse {
  workspace: Workspace
}

export interface DeleteWorkspaceRequest {
  id: string
}

export interface DeleteWorkspaceResponse {
  status: string
}
