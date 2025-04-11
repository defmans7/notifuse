import { useState, useRef } from 'react'
import { Alert, Button, Space, Tabs } from 'antd'
import { DesktopOutlined, MobileOutlined, EditOutlined } from '@ant-design/icons'
import mjml2html from 'mjml-browser'
import { Liquid } from 'liquidjs'
import { usePrismjs } from './Widgets/PrismJS'

import { DesktopWidth, MobileWidth } from './Layout'
import Iframe from './Widgets/Iframe'

import 'prismjs/components/prism-xml-doc'
import { treeToMjml } from '../utils'

interface PreviewProps {
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
  const mjml = treeToMjml(
    props.tree.data.styles,
    props.tree,
    props.templateData,
    props.urlParams,
    undefined
  )
  // console.log('mjml', mjml)
  // const mjmlBody = Prism.highlight(mjml, Prism.languages.xml, 'xml')
  const html = mjml2html(mjml)

  const iframeProps = {
    content: html.html,
    style: {
      width: props.isMobile ? '400px' : '100%',
      height: '100%',
      margin: '0 auto 0 auto',
      display: 'block',
      transition: 'all 0.1s'
    },
    sizeSelector: '.ant-drawer-body',
    id: 'htmlCompiled'
  }

  let templateError
  let jsonData = {}
  if (props.templateData && props.templateData !== '') {
    jsonData = JSON.parse(props.templateData)
  }

  try {
    const engine = new Liquid()
    // Check if html.html contains any Nunjucks filter syntax that might be incompatible with Liquid
    let templateContent = html.html
    if (templateContent.includes('|') && /\|\s*\w+\s*\(/.test(templateContent)) {
      console.warn(
        'Detected potential Nunjucks filter syntax in HTML that might be incompatible with Liquid'
      )
      // Convert Nunjucks filter syntax to Liquid filter syntax
      templateContent = templateContent.replace(/\|\s*(\w+)\s*\(([^)]*)\)/g, '|$1:$2')
      engine.parseAndRenderSync(templateContent, jsonData)
    } else {
      engine.parseAndRenderSync(html.html, jsonData)
    }
  } catch (e: any) {
    templateError = e.message
    console.error('Liquid rendering error in preview:', e)
  }

  // console.log('html', html.errors)

  // Apply syntax highlighting using usePrismjs hook
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
      {templateError && <Alert message={templateError} type="error" />}
      {tab === 'html' && (
        <div className="xpeditor-transparent" style={{ height: '100%' }}>
          <Iframe {...iframeProps} />
        </div>
      )}

      {tab === 'mjml' && (
        <div className="xpeditor-code-bg">
          {html.errors &&
            html.errors.length > 0 &&
            html.errors.map((err: any, i: number) => (
              <Alert
                key={i}
                className="xpeditor-margin-b-s"
                message={err.formattedMessage}
                type="error"
              />
            ))}
          <pre
            ref={preRef}
            className="language-xml"
            style={{
              margin: '0',
              borderRadius: '4px',
              padding: 'O',
              fontSize: '12px',
              wordWrap: 'break-word',
              whiteSpace: 'pre-wrap',
              wordBreak: 'normal'
            }}
          >
            <code className="language-xml">{mjml}</code>
          </pre>
        </div>
      )}
    </>
  )
}

export default Preview
