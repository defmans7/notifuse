import React, { useState, useEffect, useRef } from 'react'
import { Popover, Typography, Spin, Alert, Tabs } from 'antd'
import type { Template, MjmlCompileError } from '../../services/api/types'
import { templatesApi } from '../../services/api/template'
import { BlockInterface } from '../../components/email_editor/Block' // Assuming BlockInterface is here
import { usePrismjs } from '../../components/email_editor/UI/Widgets/PrismJS'
// We don't need the usePrismjs hook if we call Prism directly
// import { usePrismjs } from './email_editor/UI/Widgets/PrismJS'

const { Text } = Typography

interface TemplatePreviewPopoverProps {
  record: Template
  workspaceId: string
  children: React.ReactNode
}

const TemplatePreviewPopover: React.FC<TemplatePreviewPopoverProps> = ({
  record,
  workspaceId,
  children
}) => {
  const [previewHtml, setPreviewHtml] = useState<string | null>(null)
  const [previewMjml, setPreviewMjml] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [mjmlError, setMjmlError] = useState<MjmlCompileError | null>(null)
  const [isOpen, setIsOpen] = useState<boolean>(false)
  const [activeTabKey, setActiveTabKey] = useState<string>('1') // State for active tab

  const preRef = useRef<HTMLPreElement>(null)
  usePrismjs(preRef, ['line-numbers'])

  // Removed usePrismjs hook call

  const fetchPreview = async () => {
    if (!workspaceId || !record.email?.visual_editor_tree) {
      setError('Missing workspace ID or template data.')
      setMjmlError(null)
      setPreviewMjml(null)
      setPreviewHtml(null)
      return
    }

    setIsLoading(true)
    setError(null)
    setMjmlError(null)
    setPreviewHtml(null)
    setPreviewMjml(null)
    setActiveTabKey('1') // Reset to HTML tab on new fetch

    try {
      let treeObject: BlockInterface | null = null
      if (record.email?.visual_editor_tree && typeof record.email.visual_editor_tree === 'string') {
        try {
          treeObject = JSON.parse(record.email.visual_editor_tree)
        } catch (parseError) {
          console.error('Failed to parse visual_editor_tree:', parseError)
          setError('Invalid template structure data.')
          setMjmlError(null)
          setPreviewMjml(null)
          setIsLoading(false)
          return
        }
      } else if (record.email?.visual_editor_tree) {
        treeObject = record.email.visual_editor_tree as unknown as BlockInterface
      }

      if (!treeObject) {
        setError('Template structure data is missing or invalid.')
        setMjmlError(null)
        setPreviewMjml(null)
        setIsLoading(false)
        return
      }

      const req = {
        workspace_id: workspaceId,
        visual_editor_tree: treeObject as any,
        test_data: record.test_data || {}
      }
      // console.log('Compile Request:', req)
      const response = await templatesApi.compile(req)
      // console.log('Compile Response:', response)

      if (response.error) {
        setMjmlError(response.error)
        setPreviewMjml(response.mjml)
        setError(null)
        setPreviewHtml(null)
      } else {
        setPreviewHtml(response.html)
        setPreviewMjml(response.mjml)
        setError(null)
        setMjmlError(null)
      }
    } catch (err: any) {
      console.error('Compile Error:', err)
      const errorMsg =
        err.response?.data?.error || err.message || 'Failed to compile template preview.'
      setError(errorMsg)
      setMjmlError(null)
      setPreviewMjml(null)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (isOpen && workspaceId) {
      fetchPreview()
    } else if (!isOpen) {
      // Reset state when popover closes to avoid showing stale data briefly on reopen
      setPreviewHtml(null)
      setPreviewMjml(null)
      setError(null)
      setMjmlError(null)
      setIsLoading(false)
      setActiveTabKey('1')
    }
  }, [isOpen, record.id, record.version, workspaceId]) // Keep original dependencies

  const items = []

  if (previewHtml) {
    items.push({
      key: '1',
      label: 'HTML Preview',
      children: (
        <iframe
          srcDoc={previewHtml}
          className="w-full h-full border-0"
          style={{ height: '600px', width: '100%' }}
          title={`HTML Preview of ${record.name}`}
          sandbox="allow-same-origin"
        />
      )
    })
  }

  if (previewMjml) {
    items.push({
      key: '2',
      label: 'MJML Source',
      children: <MJMLPreview previewMjml={previewMjml} />
    })
  }
  const content = (
    <div className="w-[460px]">
      {/* Header details */}
      <div className="mb-4 space-y-2">
        <div>
          <Text strong>From: </Text>
          <Text>{record.email?.from_name}</Text>
          <Text type="secondary"> &lt;{record.email?.from_address}&gt;</Text>
        </div>
        {record.email?.reply_to && (
          <div>
            <Text strong>Reply to: </Text>
            <Text type="secondary">{record.email.reply_to}</Text>
          </div>
        )}
        <div>
          <Text strong>Subject: </Text>
          <Text>{record.email?.subject}</Text>
        </div>
        {record.email?.subject_preview && (
          <div>
            <Text strong>Preview: </Text>
            <Text type="secondary">{record.email.subject_preview}</Text>
          </div>
        )}
      </div>
      {/* Main content area */}
      <div className="flex flex-col">
        {isLoading && (
          <div className="flex items-center justify-center flex-grow">
            <Spin size="large" />
          </div>
        )}
        {!isLoading &&
          error &&
          !mjmlError && ( // General error (not MJML compilation error)
            <div className="p-4">
              <Alert message="Error loading preview" description={error} type="error" showIcon />
            </div>
          )}
        {!isLoading && mjmlError && (
          // MJML Compilation Error
          <div className="p-4 overflow-auto flex-grow flex flex-col">
            <Alert
              message={`MJML Compilation Error: ${mjmlError.message}`}
              type="error"
              showIcon
              description={
                mjmlError.details && mjmlError.details.length > 0 ? (
                  <ul className="list-disc list-inside mt-2 text-xs">
                    {mjmlError.details.map((detail, index) => (
                      <li key={index}>
                        Line {detail.line} ({detail.tagName}): {detail.message}
                      </li>
                    ))}
                  </ul>
                ) : (
                  'No specific details provided.'
                )
              }
              className="mb-4 flex-shrink-0" // Prevent alert from growing too large
            />
          </div>
        )}
        {!isLoading &&
          items.length > 0 && ( // Success case
            <Tabs
              activeKey={activeTabKey} // Control active tab
              onChange={setActiveTabKey} // Update state on tab change (onChange is preferred over onTabClick for controlled Tabs)
              className="flex flex-col flex-grow"
              items={items}
              destroyInactiveTabPane={false}
            />
          )}
        {!isLoading &&
          !error &&
          !mjmlError &&
          !previewHtml &&
          !previewMjml && ( // Neither success nor error, initial or no data state
            <div className="flex items-center justify-center flex-grow text-gray-500">
              No preview available or template is empty.
            </div>
          )}
      </div>
    </div>
  )

  return (
    <Popover
      placement="left"
      content={content}
      trigger="click"
      open={isOpen}
      onOpenChange={setIsOpen}
      // Destroy popover contents when closed to ensure clean state on reopen
      destroyTooltipOnHide={true}
    >
      {children}
    </Popover>
  )
}

const MJMLPreview = ({ previewMjml }: { previewMjml: string }) => {
  const preRef = useRef<HTMLPreElement>(null)
  usePrismjs(preRef, ['line-numbers'])

  return (
    <div className="overflow-auto">
      <pre
        ref={preRef}
        className="language-xml"
        style={{
          fontSize: '12px',
          margin: 0,
          padding: '10px'
        }}
      >
        <code className="language-xml">{previewMjml}</code>
      </pre>
    </div>
  )
}

export default TemplatePreviewPopover
