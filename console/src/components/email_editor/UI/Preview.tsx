import { useState, useRef, useEffect } from 'react'
import { Alert, Button, Space, Tabs, Spin } from 'antd'
import { DesktopOutlined, MobileOutlined, EditOutlined } from '@ant-design/icons'
import { usePrismjs } from './Widgets/PrismJS'

import { DesktopWidth, MobileWidth } from './Layout'
import Iframe from './Widgets/Iframe'
import { templatesApi } from '../../../services/api/template'
import type { MjmlCompileError } from '../../../services/api/types'

import 'prismjs/components/prism-xml-doc'

interface PreviewProps {
  workspaceId: string
  tree: any
  templateData: string
  isMobile: boolean
  deviceWidth: number
  urlParams: any
  toggleDevice: () => void
  closePreview: () => void
  onSave?: () => void
  onBack?: () => void
}

const Preview = (props: PreviewProps) => {
  const [tab, setTab] = useState('html')
  const [compiledHtml, setCompiledHtml] = useState<string | null>(null)
  const [compiledMjml, setCompiledMjml] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [apiError, setApiError] = useState<string | null>(null)
  const [mjmlError, setMjmlError] = useState<MjmlCompileError | null>(null)

  useEffect(() => {
    const compileTemplate = async () => {
      if (!props.workspaceId || !props.tree) {
        setApiError('Missing workspace ID or template tree.')
        setCompiledHtml(null)
        setCompiledMjml(null)
        setMjmlError(null)
        setIsLoading(false)
        return
      }

      setIsLoading(true)
      setApiError(null)
      setMjmlError(null)
      setCompiledHtml(null)
      setCompiledMjml(null)

      let testDataJson = {}
      if (props.templateData && props.templateData.trim() !== '') {
        try {
          testDataJson = JSON.parse(props.templateData)
        } catch (e) {
          console.error('Invalid template data JSON:', e)
          setApiError('Invalid Test Data JSON. Please check the syntax in the Test Data panel.')
          setIsLoading(false)
          setCompiledMjml(null)
          setCompiledHtml(null)
          return
        }
      }

      try {
        const req = {
          workspace_id: props.workspaceId,
          visual_editor_tree: props.tree,
          test_data: testDataJson
        }
        const response = await templatesApi.compile(req)

        if (response.error) {
          setMjmlError(response.error)
          setCompiledMjml(response.mjml)
          setApiError(null)
          setCompiledHtml(null)
          setTab('mjml')
        } else {
          setCompiledHtml(response.html)
          setCompiledMjml(response.mjml)
          setApiError(null)
          setMjmlError(null)
        }
      } catch (err: any) {
        console.error('API Compile Error in Preview:', err)
        const errorMsg =
          err.response?.data?.error || err.message || 'Failed to compile template preview.'
        setApiError(errorMsg)
        setMjmlError(null)
        setCompiledMjml(null)
        setCompiledHtml(null)
      } finally {
        setIsLoading(false)
      }
    }

    compileTemplate()
  }, [props.tree, props.templateData, props.workspaceId, props.urlParams])

  const iframeProps = {
    content: compiledHtml || '',
    style: {
      width: props.isMobile ? MobileWidth + 'px' : '100%',
      height: '100%',
      margin: '0 auto 0 auto',
      display: 'block',
      border: 'none',
      transition: 'width 0.3s ease-in-out'
    },
    sizeSelector: '.ant-drawer-body',
    id: 'visual-editor-preview-iframe'
  }

  const preRef = useRef<HTMLPreElement>(null)
  usePrismjs(preRef, ['line-numbers'])

  return (
    <>
      <div className="xpeditor-topbar">
        <span style={{ float: 'right' }}>
          <Space>
            <Space.Compact>
              <Button
                size="small"
                type="text"
                disabled={props.deviceWidth === MobileWidth}
                onClick={() => props.toggleDevice()}
              >
                <MobileOutlined />
              </Button>
              <Button
                size="small"
                type="text"
                disabled={props.deviceWidth === DesktopWidth}
                onClick={() => props.toggleDevice()}
              >
                <DesktopOutlined />
              </Button>
            </Space.Compact>
            <Button type="primary" size="small" ghost onClick={() => props.closePreview()}>
              <EditOutlined />
              &nbsp; Edit
            </Button>
          </Space>
        </span>
        <Tabs
          activeKey={tab}
          centered
          onChange={(k) => setTab(k)}
          style={{ position: 'absolute', top: '6px' }}
          items={[
            {
              key: 'html',
              label: 'HTML'
            },
            {
              key: 'mjml',
              label: 'MJML'
            }
          ]}
        />
      </div>

      {isLoading && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: 'calc(100% - 50px)'
          }}
        >
          <Spin size="large" tip="Compiling Preview..." />
        </div>
      )}

      {!isLoading && apiError && (
        <Alert
          message="Preview Error"
          description={apiError}
          type="error"
          showIcon
          style={{ margin: '10px' }}
        />
      )}

      {!isLoading && tab === 'html' && !apiError && (
        <div
          className="xpeditor-transparent"
          style={{ height: 'calc(100% - 50px)', overflow: 'hidden' }}
        >
          {compiledHtml ? (
            <Iframe {...iframeProps} />
          ) : (
            <div style={{ padding: '20px', textAlign: 'center' }}>
              {mjmlError ? 'HTML preview unavailable due to MJML errors.' : 'No HTML generated.'}
            </div>
          )}
        </div>
      )}

      {!isLoading && tab === 'mjml' && !apiError && (
        <div
          className="xpeditor-code-bg"
          style={{ height: 'calc(100% - 50px)', overflowY: 'auto' }}
        >
          {mjmlError && (
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
              style={{ margin: '10px 10px 0 10px' }}
            />
          )}
          {compiledMjml ? (
            <pre
              ref={preRef}
              key={compiledMjml}
              className="language-xml line-numbers"
              style={{
                margin: '10px',
                borderRadius: '4px',
                padding: '10px',
                fontSize: '12px',
                wordWrap: 'break-word',
                whiteSpace: 'pre-wrap',
                wordBreak: 'normal'
              }}
            >
              <code className="language-xml">{compiledMjml}</code>
            </pre>
          ) : (
            <div style={{ padding: '20px', textAlign: 'center' }}>
              {mjmlError
                ? 'MJML source unavailable due to compilation errors.'
                : 'No MJML source available.'}
            </div>
          )}
        </div>
      )}
    </>
  )
}

export default Preview
