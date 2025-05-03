import { useEffect, useState } from 'react'
import { Button, Form, Input, Select, App } from 'antd'
import { Workspace } from '../../services/api/types'
import { workspaceService } from '../../services/api/workspace'
import { useNavigate } from '@tanstack/react-router'
import { Section } from './Section'
import { TimezonesFormOptions } from '../utils/countries_timezones'

const { Option } = Select

interface WorkspaceSettingsProps {
  workspace: Workspace | null
  loading: boolean
  onWorkspaceUpdate: (workspace: Workspace) => void
  onWorkspaceDelete?: () => void
  isOwner: boolean
}

export function WorkspaceSettings({ workspace, onWorkspaceUpdate }: WorkspaceSettingsProps) {
  const [savingSettings, setSavingSettings] = useState(false)
  const [formTouched, setFormTouched] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  useEffect(() => {
    // Set form values from workspace data whenever workspace changes
    form.setFieldsValue({
      name: workspace?.name || '',
      website_url: workspace?.settings.website_url || '',
      timezone: workspace?.settings.timezone || 'UTC'
    })
    setFormTouched(false)
  }, [workspace, form])

  const handleSaveSettings = async (values: any) => {
    if (!workspace) return

    setSavingSettings(true)
    try {
      await workspaceService.update({
        id: workspace.id,
        name: values.name,
        settings: {
          website_url: values.website_url,
          logo_url: workspace?.settings.logo_url || null,
          cover_url: workspace?.settings.cover_url || null,
          timezone: values.timezone
        }
      })

      // Refresh the workspace data
      const response = await workspaceService.get(workspace.id)

      // Update the parent component with the new workspace data
      onWorkspaceUpdate(response.workspace)

      setFormTouched(false)
      message.success('Workspace settings updated successfully')
    } catch (error) {
      console.error('Failed to update workspace settings', error)
      message.error('Failed to update workspace settings')
    } finally {
      setSavingSettings(false)
    }
  }

  const handleFormChange = () => {
    setFormTouched(true)
  }

  return (
    <>
      <Section title="General Settings" description="General settings for your workspace">
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveSettings}
          onValuesChange={handleFormChange}
        >
          <Form.Item
            name="name"
            label="Workspace Name"
            rules={[{ required: true, message: 'Please enter workspace name' }]}
          >
            <Input placeholder="Enter workspace name" />
          </Form.Item>

          <Form.Item
            name="website_url"
            label="Website URL"
            rules={[{ type: 'url', message: 'Please enter a valid URL' }]}
          >
            <Input placeholder="https://example.com" />
          </Form.Item>

          <Form.Item
            name="timezone"
            label="Timezone"
            rules={[{ required: true, message: 'Please select a timezone' }]}
          >
            <Select options={TimezonesFormOptions} showSearch optionFilterProp="label" />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={savingSettings}
              disabled={!formTouched}
            >
              Save Changes
            </Button>
          </Form.Item>
        </Form>
      </Section>
    </>
  )
}
