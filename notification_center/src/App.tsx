import { useState, useEffect } from 'react'
import { getNotificationCenter, parseNotificationCenterParams } from './api/notification_center'
import type { NotificationCenterResponse } from './api/notification_center'
import { Switch } from './components/ui/switch'

function App() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [notificationData, setNotificationData] = useState<NotificationCenterResponse | null>(null)
  const [subscriptions, setSubscriptions] = useState<Record<string, boolean>>({})

  useEffect(() => {
    async function loadNotificationData() {
      try {
        // Parse parameters from URL
        const params = parseNotificationCenterParams()

        if (!params) {
          setError('Missing required parameters. Please check the URL.')
          setLoading(false)
          return
        }

        // Load notification center data
        const data = await getNotificationCenter(params)
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

  const handleSubscriptionToggle = (listId: string) => {
    setSubscriptions((prev) => ({
      ...prev,
      [listId]: !prev[listId]
    }))

    // In a real implementation, you would call an API to update the subscription
    console.log(`Toggled subscription for list ${listId} to ${!subscriptions[listId]}`)
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
          <div className="text-base md:text-lg font-medium text-gray-800 flex-1 text-center">
            <a
              href={websiteUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:underline"
            >
              Notification Center
            </a>
          </div>
          {/* Empty div for flex balance */}
          <div className="flex-shrink-0 w-8 md:w-10"></div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col items-center p-4">
        <div className="w-full max-w-[600px]">
          <div className="p-4">
            {notificationData && (
              <>
                <div className="mb-6">
                  <h2 className="text-lg font-medium">
                    Welcome, {notificationData.contact.first_name || notificationData.contact.email}
                  </h2>
                </div>

                {/* Merged Lists section with toggles */}
                {publicLists.length > 0 && (
                  <div className="mb-6">
                    <h3 className="font-medium mb-3 text-gray-700">Email Subscriptions</h3>
                    <div className="space-y-3">
                      {publicLists.map((list) => {
                        const isSubscribed = subscriptions[list.id] || false

                        return (
                          <div key={list.id} className="p-4 border rounded-lg bg-white">
                            <div className="flex items-center justify-between">
                              <div className="flex-1">
                                <div className="font-medium">{list.name}</div>
                                {list.description && (
                                  <p className="text-sm text-gray-600 mt-1">{list.description}</p>
                                )}
                              </div>
                              <div className="ml-4">
                                <Switch
                                  checked={isSubscribed}
                                  onCheckedChange={() => handleSubscriptionToggle(list.id)}
                                />
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
