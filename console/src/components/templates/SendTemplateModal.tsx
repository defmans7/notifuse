import { useState, useEffect } from 'react'
import { Modal, Button, Input, Select, Typography, Space } from 'antd'
import { Workspace, Template, Integration } from '../../services/api/types'
import { emailService } from '../../services/api/email'
import { message } from 'antd'
import { emailProviders } from '../integrations/EmailProviders'

const { Text } = Typography
const { Option } = Select

interface SendTemplateModalProps {
  isOpen: boolean
  onClose: () => void
  template: Template | null
  workspace: Workspace | null
  loading?: boolean
}

export default function SendTemplateModal({
  isOpen,
  onClose,
  template,
  workspace,
  loading = false
}: SendTemplateModalProps) {
  const [email, setEmail] = useState('')
  const [selectedIntegrationId, setSelectedIntegrationId] = useState<string>('')
  const [sendLoading, setSendLoading] = useState(false)

  // Filter to only email integrations
  const emailIntegrations =
    workspace?.integrations?.filter(
      (integration) => integration.type === 'email' && integration.email_provider?.kind
    ) || []

  // Set default integration when modal opens or template changes
  useEffect(() => {
    if (isOpen && workspace && emailIntegrations.length > 0 && !selectedIntegrationId) {
      const defaultId =
        template?.category === 'marketing'
          ? workspace.settings?.marketing_email_provider_id
          : workspace.settings?.transactional_email_provider_id

      // Use the appropriate default or the first available integration
      setSelectedIntegrationId(
        defaultId && emailIntegrations.some((i) => i.id === defaultId)
          ? defaultId
          : emailIntegrations[0]?.id || ''
      )
    }
  }, [isOpen, template, workspace, emailIntegrations, selectedIntegrationId])

  const handleSend = async () => {
    if (!template || !workspace || !selectedIntegrationId) return

    setSendLoading(true)
    try {
      const response = await emailService.testTemplate(
        workspace.id,
        template.id,
        selectedIntegrationId,
        email
      )

      if (response.success) {
        message.success('Test email sent successfully')
        onClose()
        setEmail('')
      } else {
        message.error(`Failed to send test email: ${response.error || 'Unknown error'}`)
      }
    } catch (error: any) {
      message.error(`Error: ${error?.message || 'Something went wrong'}`)
    } finally {
      setSendLoading(false)
    }
  }

  const renderIntegrationOption = (integration: Integration) => {
    const providerKind = integration.email_provider.kind
    // Find the provider info to get the icon
    const providerInfo = emailProviders.find((p) => p.kind === providerKind)

    return (
      <Option key={integration.id} value={integration.id}>
        <span className="mr-1">
          {providerInfo ? providerInfo.getIcon('mr-1') : <span className="h-5 w-5 inline-block" />}
        </span>
        <span>{integration.name}</span>
      </Option>
    )
  }

  return (
    <Modal
      title="Send Test Email"
      open={isOpen}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="send"
          type="primary"
          onClick={handleSend}
          disabled={!email || !selectedIntegrationId || loading || sendLoading}
          loading={loading || sendLoading}
        >
          Send Test Email
        </Button>
      ]}
    >
      <div className="py-2 space-y-4">
        <p>Send a test email using this template to verify how it will look.</p>

        <div>
          <div className="mb-1">Email Integration</div>
          <Select
            className="w-full"
            placeholder="Select an email integration"
            value={selectedIntegrationId}
            onChange={setSelectedIntegrationId}
            disabled={emailIntegrations.length === 0}
          >
            {emailIntegrations.map(renderIntegrationOption)}
          </Select>
          {emailIntegrations.length === 0 && (
            <Text type="warning" className="mt-1 block">
              No email integrations available. Please configure one in Settings.
            </Text>
          )}
        </div>

        <div>
          <div className="mb-1">Recipient Email</div>
          <Input
            placeholder="recipient@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            type="email"
          />
        </div>
      </div>
    </Modal>
  )
}
