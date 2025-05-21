import { useState, useEffect } from 'react'
import {
  getContactPreferences,
  parseNotificationCenterParams,
  subscribeToLists,
  unsubscribeOneClick
} from './api/notification_center'
import type { ContactPreferencesResponse, List } from './api/notification_center'
import { Button } from '@/components/ui/button'
import { Toaster } from '@/components/ui/sonner'
import { toast } from 'sonner'

function App() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [notificationData, setNotificationData] = useState<ContactPreferencesResponse | null>(null)
  const [subscriptions, setSubscriptions] = useState<Record<string, boolean>>({})
  const [processingLists, setProcessingLists] = useState<Record<string, boolean>>({})
  const [allLists, setAllLists] = useState<Array<List & { status?: string }>>([])

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

        // Combine public lists and contact-specific lists
        const combinedLists: Array<List & { status?: string }> = []

        // Process all contact lists to get status
        if (data.contact_lists) {
          data.contact_lists.forEach((contactList) => {
            // Set subscription status based on contact list status
            initialSubscriptions[contactList.list_id] = contactList.status === 'active'

            // Try to find this list in public lists to get name and description
            const publicList = data.public_lists?.find((list) => list.id === contactList.list_id)

            if (publicList) {
              // For lists in both contact_lists and public_lists
              combinedLists.push({
                ...publicList,
                status: contactList.status
              })
            } else {
              // For lists only in contact_lists (private lists)
              combinedLists.push({
                id: contactList.list_id,
                name: contactList.list_name || `List ${contactList.list_id}`,
                status: contactList.status
              })
            }
          })
        }

        // Add public lists that aren't in contact_lists
        if (data.public_lists) {
          data.public_lists.forEach((list) => {
            const existingList = combinedLists.find((l) => l.id === list.id)
            if (!existingList) {
              combinedLists.push({
                ...list,
                status: 'unsubscribed' // Default status for public lists not in contact_lists
              })
              initialSubscriptions[list.id] = false
            }
          })
        }

        setAllLists(combinedLists)
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

      // Also update list status if possible
      setAllLists((prev) =>
        prev.map((list) => (list.id === listId ? { ...list, status: 'active' } : list))
      )

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

      // Revert list status
      setAllLists((prev) =>
        prev.map((list) => (list.id === listId ? { ...list, status: 'unsubscribed' } : list))
      )

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

      // Also update list status if possible
      setAllLists((prev) =>
        prev.map((list) => (list.id === listId ? { ...list, status: 'unsubscribed' } : list))
      )

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

      // Revert list status
      setAllLists((prev) =>
        prev.map((list) => (list.id === listId ? { ...list, status: 'active' } : list))
      )

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

              {/* All Lists section with toggles - showing both public and private lists */}
              {allLists.length > 0 && (
                <div className="mb-6">
                  <div className="space-y-3">
                    {allLists.map((list) => {
                      const isSubscribed = subscriptions[list.id] || false
                      const isActive = list.status === 'active'
                      const canToggle = list.status !== 'bounced' && list.status !== 'complained'

                      return (
                        <div
                          key={list.id}
                          className={`p-4 border border-gray-300 rounded-lg ${
                            isActive ? 'bg-white' : 'bg-gray-50'
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex-1">
                              <div className="font-medium">
                                {list.name}
                                {list.status &&
                                  list.status !== 'active' &&
                                  list.status !== 'unsubscribed' && (
                                    <span className="ml-2 text-xs px-2 py-1 bg-gray-200 text-gray-700 rounded-full">
                                      {list.status}
                                    </span>
                                  )}
                              </div>
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
                                disabled={processingLists[list.id] || !canToggle}
                                className={`cursor-pointer ${
                                  !canToggle
                                    ? 'border-gray-300 text-gray-400 cursor-not-allowed'
                                    : isSubscribed
                                    ? 'border-red-500 text-red-500 hover:bg-red-50'
                                    : 'border-blue-500 text-blue-500 hover:bg-blue-50'
                                }`}
                              >
                                {processingLists[list.id]
                                  ? 'Processing...'
                                  : !canToggle
                                  ? list.status
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
              {allLists.length === 0 && (
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
