import { api } from './client'

export interface NotificationCenterParams {
  workspace_id: string
  email: string
  email_hmac: string
}

export interface Contact {
  id: string
  email: string
  first_name?: string
  last_name?: string
  external_id?: string
  [key: string]: any
}

export interface List {
  id: string
  name: string
  description?: string
}

export interface ContactList {
  list_id: string
  subscribed: boolean
}

export interface TransactionalNotification {
  id: string
  name: string
  description?: string
}

export interface NotificationCenterResponse {
  contact: Contact
  public_lists?: List[] | null
  contact_lists?: ContactList[] | null
  logo_url?: string
  website_url?: string
}

/**
 * Validates notification center parameters
 * Matches the Validate method in NotificationCenterRequest
 */
export function validateParams(params: Partial<NotificationCenterParams>): string | null {
  if (!params.email) {
    return 'email is required'
  }
  if (!params.email_hmac) {
    return 'email_hmac is required'
  }
  if (!params.workspace_id) {
    return 'workspace_id is required'
  }
  return null
}

/**
 * Fetches notification center data for a contact
 * This uses the public endpoint that doesn't require authentication
 * Matches the handleNotificationCenter method in ContactHandler
 */
export async function getNotificationCenter(
  params: NotificationCenterParams
): Promise<NotificationCenterResponse> {
  // Validate parameters first
  const validationError = validateParams(params)
  if (validationError) {
    throw new Error(validationError)
  }

  const queryParams = new URLSearchParams({
    workspace_id: params.workspace_id,
    email: params.email,
    email_hmac: params.email_hmac
  }).toString()

  return api.get<NotificationCenterResponse>(`/notification-center?${queryParams}`)
}

/**
 * Marks a message as read
 * This would need a corresponding endpoint in the backend
 */
export async function markMessageAsRead(
  messageId: string,
  params: NotificationCenterParams
): Promise<void> {
  // Validate parameters first
  const validationError = validateParams(params)
  if (validationError) {
    throw new Error(validationError)
  }

  return api.post<void>('/notification-center/mark-read', {
    message_id: messageId,
    workspace_id: params.workspace_id,
    email: params.email,
    email_hmac: params.email_hmac
  })
}

/**
 * Utility function to parse notification center parameters from URL
 * Similar to FromURLValues method in NotificationCenterRequest
 */
export function parseNotificationCenterParams(): NotificationCenterParams | null {
  const searchParams = new URLSearchParams(window.location.search)

  const params: Partial<NotificationCenterParams> = {
    workspace_id: searchParams.get('workspace_id') || undefined,
    email: searchParams.get('email') || undefined,
    email_hmac: searchParams.get('email_hmac') || undefined
  }

  // Check if all required params are present
  const validationError = validateParams(params)
  if (validationError) {
    return null
  }

  return params as NotificationCenterParams
}
