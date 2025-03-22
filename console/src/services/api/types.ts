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
  website_url: string
  logo_url: string | null
  cover_url: string | null
  timezone: string
}

export interface CreateWorkspaceRequest {
  id: string
  name: string
  settings: WorkspaceSettings
}

export interface Workspace {
  id: string
  name: string
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
  name?: string
  settings?: {
    website_url?: string
    logo_url?: string | null
    cover_url?: string | null
    timezone?: string
  }
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

// Workspace Member types
export interface WorkspaceMember {
  user_id: string
  workspace_id: string
  role: string
  email: string
  created_at: string
  updated_at: string
}

export interface GetWorkspaceMembersResponse {
  members: WorkspaceMember[]
}

// Workspace Member Invitation types
export interface InviteMemberRequest {
  workspace_id: string
  email: string
  role: string
}

export interface InviteMemberResponse {
  status: string
  message: string
}
