import { useState, useEffect } from 'react'
import {
  getContactPreferences,
  parseNotificationCenterParams,
  subscribeToLists,
  unsubscribeOneClick
} from './api/notification_center'
import type { ContactPreferencesResponse } from './api/notification_center'
import { Button } from '@/components/ui/button'
import { Toaster } from '@/components/ui/sonner'
import { toast } from 'sonner'

function App() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [notificationData, setNotificationData] = useState<ContactPreferencesResponse | null>(null)
  const [subscriptions, setSubscriptions] = useState<Record<string, boolean>>({})
  const [processingLists, setProcessingLists] = useState<Record<string, boolean>>({})

  useEffect(() => {
    async function loadNotificationData() {
      ;(window as any).Consecration =
        'Imperet illi Deus, supplices deprecamur: tuque, Princeps militiae caelestis, Satanam aliosque spiritus malignos, qui ad perditionem animarum pervagantur in mundo, divina virtute in infernum detrude. Amen.'

      try {
        // Parse parameters from URL
        const params = parseNotificationCenterParams()

        if (!params) {
          setError('Missing required parameters. Please check the URL.')
          setLoading(false)
          return
        }

        // Load notification center data
        const data = await getContactPreferences({
          workspace_id: params.wid,
          email: params.email,
          email_hmac: params.email_hmac
        })
        setNotificationData(data)

        // Initialize subscriptions state
        const initialSubscriptions: Record<string, boolean> = {}
        if (data.contact_lists && data.public_lists) {
          data.public_lists.forEach((list) => {
            const contactList = data.contact_lists?.find((cl) => cl.list_id === list.id)
            initialSubscriptions[list.id] = contactList?.subscribed || false
          })
        }
        setSubscriptions(initialSubscriptions)

        setLoading(false)
      } catch (err) {
        console.error('Failed to load notification center data:', err)
        setError(err instanceof Error ? err.message : 'Failed to load notifications')
        setLoading(false)
      }
    }

    loadNotificationData()
  }, [])

  // Set favicon when logo is available
  useEffect(() => {
    if (notificationData?.logo_url) {
      const existingLink = document.querySelector("link[rel*='icon']") as HTMLLinkElement | null
      const link = existingLink || document.createElement('link')
      link.type = 'image/x-icon'
      link.rel = 'shortcut icon'
      link.href = notificationData.logo_url

      if (!existingLink) {
        document.head.appendChild(link)
      }
    }
  }, [notificationData?.logo_url])

  // Update page title with contact information
  useEffect(() => {
    if (notificationData?.contact) {
      document.title = `${notificationData.contact.email} | Email Subscriptions`
    }
  }, [notificationData?.contact])

  const subscribe = async (listId: string) => {
    try {
      // Set processing state
      setProcessingLists((prev) => ({ ...prev, [listId]: true }))

      // Update local state optimistically
      setSubscriptions((prev) => ({
        ...prev,
        [listId]: true
      }))

      if (notificationData?.contact) {
        const params = parseNotificationCenterParams()

        if (!params) {
          throw new Error('Missing required parameters')
        }

        // Call API to subscribe to list
        await subscribeToLists({
          workspace_id: params.wid,
          contact: notificationData.contact,
          list_ids: [listId]
        })

        toast.success('Successfully subscribed', {
          style: { backgroundColor: '#f0fdf4', borderLeft: '4px solid #22c55e', color: '#166534' },
          duration: 3000
        })
      }
    } catch (err) {
      // Revert local state on error
      setSubscriptions((prev) => ({
        ...prev,
        [listId]: false
      }))

      console.error('Failed to subscribe:', err)
      toast.error('Failed to subscribe. Please try again.')
    } finally {
      // Clear processing state
      setProcessingLists((prev) => ({ ...prev, [listId]: false }))
    }
  }

  const unsubscribe = async (listId: string) => {
    try {
      // Set processing state
      setProcessingLists((prev) => ({ ...prev, [listId]: true }))

      // Update local state optimistically
      setSubscriptions((prev) => ({
        ...prev,
        [listId]: false
      }))

      if (notificationData?.contact) {
        const params = parseNotificationCenterParams()

        if (!params) {
          throw new Error('Missing required parameters')
        }

        // Call API to unsubscribe from list
        await unsubscribeOneClick({
          wid: params.wid,
          email: params.email,
          email_hmac: params.email_hmac,
          lids: [listId],
          mid: params.mid
        })

        toast.success('Successfully unsubscribed', {
          style: { backgroundColor: '#f0fdf4', borderLeft: '4px solid #22c55e', color: '#166534' },
          duration: 3000
        })
      }
    } catch (err) {
      // Revert local state on error
      setSubscriptions((prev) => ({
        ...prev,
        [listId]: true
      }))

      console.error('Failed to unsubscribe:', err)
      toast.error('Failed to unsubscribe. Please try again.')
    } finally {
      // Clear processing state
      setProcessingLists((prev) => ({ ...prev, [listId]: false }))
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-white">
        <div className="p-6 max-w-sm mx-auto">
          <div className="text-center">
            <div className="text-xl font-medium text-black">Loading...</div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-white">
        <div className="p-6 max-w-sm mx-auto">
          <div className="text-center">
            <div className="text-xl font-medium text-red-500">Error</div>
            <p className="text-gray-700 mt-2">{error}</p>
          </div>
        </div>
      </div>
    )
  }

  // Safe array access with defaults
  const publicLists = notificationData?.public_lists || []
  const websiteUrl = notificationData?.website_url || '#'

  return (
    <div className="min-h-screen flex flex-col bg-white">
      <Toaster />
      {/* Topbar with bottom border */}
      <div className="bg-white border-b border-gray-200 w-full">
        <div className="flex items-center h-16 px-4 max-w-[600px] mx-auto">
          <div className="flex-shrink-0 mr-4 md:mr-6">
            {notificationData?.logo_url ? (
              <a href={websiteUrl} target="_blank" rel="noopener noreferrer" title="Visit website">
                <img
                  src={notificationData.logo_url}
                  alt="Workspace Logo"
                  className="h-8 md:h-10 w-auto object-contain"
                />
              </a>
            ) : (
              <div className="w-8 md:w-10 h-8 md:h-10"></div> /* Empty space when no logo */
            )}
          </div>
          <div className="text-sm font-medium text-gray-800 flex-1 text-center">
            <a
              href={websiteUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:underline"
            >
              Email Subscriptions
            </a>
          </div>
          {/* Empty div for flex balance */}
          <div className="flex-shrink-0 w-8 md:w-10"></div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col items-center p-4">
        <div className="w-full max-w-[600px]">
          {notificationData && (
            <>
              <div className="mb-6 mt-4">
                <div className="text-md font-medium">
                  Welcome, {notificationData.contact.first_name || notificationData.contact.email}
                </div>
              </div>

              {/* Merged Lists section with toggles */}
              {publicLists.length > 0 && (
                <div className="mb-6">
                  <div className="space-y-3">
                    {publicLists.map((list) => {
                      const isSubscribed = subscriptions[list.id] || false

                      return (
                        <div
                          key={list.id}
                          className="p-4 border border-gray-300 rounded-lg bg-white"
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex-1">
                              <div className="font-medium">{list.name}</div>
                              {list.description && (
                                <p className="text-sm text-gray-600 mt-1">{list.description}</p>
                              )}
                            </div>
                            <div className="ml-4">
                              <Button
                                variant="outline"
                                onClick={() =>
                                  isSubscribed ? unsubscribe(list.id) : subscribe(list.id)
                                }
                                size="sm"
                                disabled={processingLists[list.id]}
                                className={`cursor-pointer ${
                                  isSubscribed
                                    ? 'border-red-500 text-red-500 hover:bg-red-50'
                                    : 'border-blue-500 text-blue-500 hover:bg-blue-50'
                                }`}
                              >
                                {processingLists[list.id]
                                  ? 'Processing...'
                                  : isSubscribed
                                  ? 'Unsubscribe'
                                  : 'Subscribe'}
                              </Button>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}

              {/* Empty state when no lists */}
              {publicLists.length === 0 && (
                <p className="text-center text-gray-500 py-4">
                  No subscriptions settings available.
                </p>
              )}
            </>
          )}
        </div>
      </div>

      {/* Footer */}
      <div className="border-t border-gray-200 py-4 text-center text-sm text-gray-500">
        <a
          href={websiteUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-gray-700 hover:underline"
        >
          Visit our website
        </a>
      </div>
    </div>
  )
}

export default App
