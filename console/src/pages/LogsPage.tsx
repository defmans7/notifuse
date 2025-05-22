import { Typography, Tabs } from 'antd'
import { useParams } from '@tanstack/react-router'
import { MessageHistoryTab } from '../components/messages/MessageHistoryTab'
import { WebhookEventsTab } from '../components/webhooks/WebhookEventsTab'

const { Text } = Typography

export function LogsPage() {
  const { workspaceId } = useParams({ strict: false })

  if (!workspaceId) {
    return <div>Loading...</div>
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <div className="text-2xl font-medium">Logs</div>
        <Text type="secondary">Monitor message delivery status and webhook events</Text>
      </div>

      <Tabs
        defaultActiveKey="messages"
        items={[
          {
            key: 'messages',
            label: 'Message History',
            children: <MessageHistoryTab workspaceId={workspaceId} />
          },
          {
            key: 'webhooks',
            label: 'Webhooks',
            children: <WebhookEventsTab workspaceId={workspaceId} />
          }
        ]}
      />
    </div>
  )
}
