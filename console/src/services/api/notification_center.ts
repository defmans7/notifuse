import { api } from './client'
import type {
  NotificationCenterRequest,
  NotificationCenterResponse,
  SubscribeToListsRequest,
  UnsubscribeFromListsRequest
} from './types'

export const notificationCenterApi = {
  // Get notification center data for a contact
  getNotificationCenter: async (
    params: NotificationCenterRequest
  ): Promise<NotificationCenterResponse> => {
    const searchParams = new URLSearchParams()
    searchParams.append('workspace_id', params.workspace_id)
    searchParams.append('email', params.email)
    searchParams.append('email_hmac', params.email_hmac)

    return api.get<NotificationCenterResponse>(`/notification-center?${searchParams.toString()}`)
  },

  // Subscribe to lists - public route
  subscribe: async (params: SubscribeToListsRequest): Promise<{ success: boolean }> => {
    return api.post('/subscribe', params)
  },

  // One-click unsubscribe for Gmail header link
  unsubscribeOneClick: async (
    params: UnsubscribeFromListsRequest
  ): Promise<{ success: boolean }> => {
    return api.post('/unsubscribe-oneclick', params)
  }
}
