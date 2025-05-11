import { api } from './client'

export interface MessageData {
  data: Record<string, any>
  metadata?: Record<string, any>
}

export type MessageStatus =
  | 'sent'
  | 'delivered'
  | 'failed'
  | 'opened'
  | 'clicked'
  | 'bounced'
  | 'complained'
  | 'unsubscribed'

export interface MessageHistory {
  id: string
  contact_email: string
  broadcast_id?: string
  template_id: string
  template_version: number
  channel: string
  status: MessageStatus
  error?: string
  message_data: MessageData

  // Event timestamps
  sent_at: string
  delivered_at?: string
  failed_at?: string
  opened_at?: string
  clicked_at?: string
  bounced_at?: string
  complained_at?: string
  unsubscribed_at?: string

  // System timestamps
  created_at: string
  updated_at: string
}

export interface MessageListParams {
  cursor?: string
  limit?: number

  // Filters
  channel?: string
  status?: MessageStatus
  contact_email?: string
  broadcast_id?: string
  template_id?: string
  has_error?: boolean

  // Time range filters
  sent_after?: string
  sent_before?: string
  updated_after?: string
  updated_before?: string
}

export interface MessageListResult {
  messages: MessageHistory[]
  next_cursor?: string
  has_more: boolean
}

/**
 * Lists message history with pagination and filtering
 */
export function listMessages(
  workspaceId: string,
  params: MessageListParams
): Promise<MessageListResult> {
  // Convert params object to URLSearchParams for query string
  const queryParams = new URLSearchParams()
  queryParams.append('workspace_id', workspaceId)

  // Add all other params that are defined
  if (params.cursor) queryParams.append('cursor', params.cursor)
  if (params.limit) queryParams.append('limit', String(params.limit))
  if (params.channel) queryParams.append('channel', params.channel)
  if (params.status) queryParams.append('status', params.status)
  if (params.contact_email) queryParams.append('contact_email', params.contact_email)
  if (params.broadcast_id) queryParams.append('broadcast_id', params.broadcast_id)
  if (params.template_id) queryParams.append('template_id', params.template_id)
  if (params.has_error !== undefined) queryParams.append('has_error', String(params.has_error))
  if (params.sent_after) queryParams.append('sent_after', params.sent_after)
  if (params.sent_before) queryParams.append('sent_before', params.sent_before)
  if (params.updated_after) queryParams.append('updated_after', params.updated_after)
  if (params.updated_before) queryParams.append('updated_before', params.updated_before)

  return api.get<MessageListResult>(`/api/messages.list?${queryParams.toString()}`)
}
