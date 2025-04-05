import { useState, useCallback, useRef, useEffect } from 'react'
import { Form } from 'antd'
import { cloneDeep } from 'lodash'
import uuid from 'short-uuid'

import { BlockInterface, BlockDefinitionInterface } from './Block'
import ButtonBlockDefinition from './UI/definitions/Button'
import ColumnBlockDefinition from './UI/definitions/Column'
import Columns168BlockDefinition from './UI/definitions/Columns168'
import Columns204BlockDefinition from './UI/definitions/Columns204'
import Columns420BlockDefinition from './UI/definitions/Columns420'
import Columns816BlockDefinition from './UI/definitions/Columns816'
import Columns888BlockDefinition from './UI/definitions/Columns888'
import Columns1212BlockDefinition from './UI/definitions/Columns1212'
import Columns6666BlockDefinition from './UI/definitions/Columns6666'
import DividerBlockDefinition from './UI/definitions/Divider'
import HeadingBlockDefinition from './UI/definitions/Heading'
import ImageBlockDefinition from './UI/definitions/Image'
import OneColumnBlockDefinition from './UI/definitions/OneColumn'
import OpenTrackingBlockDefinition from './UI/definitions/OpenTracking'
import RootBlockDefinition from './UI/definitions/Root'
import TextBlockDefinition from './UI/definitions/Text'
import { Editor, SelectedBlockButtonsProp } from './Editor'
import { ExportHTML } from './UI/Preview'
import { Layout, DesktopWidth } from './UI/Layout'
import SelectedBlockButtons from './UI/SelectedBlockButtons'
import { FilesSettings } from '../file_manager/interfaces'
import { StrictCSSProperties, UrlParams, HtmlExportResult, EmailTemplateBlock } from './types'
import LiquidTemplateBlockDefinition from './UI/definitions/Liquid'

// Helper function to generate a block from definition
const generateBlockFromDefinition = (blockDefinition: BlockDefinitionInterface) => {
  const id = uuid.generate()

  const block: BlockInterface = {
    id: id,
    kind: blockDefinition.kind,
    path: '', // path is set when rendering
    children: blockDefinition.children
      ? blockDefinition.children.map((child: BlockDefinitionInterface) => {
          return generateBlockFromDefinition(child)
        })
      : [],
    data: cloneDeep(blockDefinition.defaultData)
  }

  return block
}

// Create default blocks
const createDefaultBlocks = () => {
  // Create default content blocks
  const text = generateBlockFromDefinition(TextBlockDefinition)
  const heading = generateBlockFromDefinition(HeadingBlockDefinition)
  const logo = generateBlockFromDefinition(ImageBlockDefinition)
  const divider = generateBlockFromDefinition(DividerBlockDefinition)
  const openTracking = generateBlockFromDefinition(OpenTrackingBlockDefinition)
  const btn = generateBlockFromDefinition(ButtonBlockDefinition)
  const column = generateBlockFromDefinition(OneColumnBlockDefinition)

  // Configure logo
  logo.data.image.src = 'https://notifuse.com/images/logo.png'
  logo.data.image.alt = 'Logo'
  logo.data.image.href = 'https://notifuse.com'
  logo.data.image.width = '100px'

  // Configure heading
  heading.data.paddingControl = 'separate'
  heading.data.paddingTop = '40px'
  heading.data.paddingBottom = '40px'
  heading.data.editorData[0].children[0].text = 'Hello {{ user.first_name | default:"there" }} 👋'

  // Configure divider
  divider.data.paddingControl = 'separate'
  divider.data.paddingTop = '40px'
  divider.data.paddingBottom = '20px'
  divider.data.paddingLeft = '200px'
  divider.data.paddingRight = '200px'

  // Configure text
  text.data.editorData[0].children[0].text = 'Welcome to the email editor!'

  // Configure button
  btn.data.button.backgroundColor = '#4e6cff'
  btn.data.button.text = '👉 Click me'

  // Add all blocks to column
  column.children[0].children.push(logo)
  column.children[0].children.push(heading)
  column.children[0].children.push(text)
  column.children[0].children.push(divider)
  column.children[0].children.push(btn)
  column.children[0].children.push(openTracking)

  // Create root block with column as child
  const rootData = cloneDeep(RootBlockDefinition.defaultData)
  const rootBlock: BlockInterface = {
    id: 'root',
    kind: 'root',
    path: '',
    children: [column],
    data: rootData
  }

  return rootBlock
}

// Define types for props
export interface DefaultEditorProps {
  onChange?: (exportedHtml: string) => void
  workspaceId?: string
  initialValue?: BlockInterface
  defaultTemplateData?: string
  blockDefinitions?: { [key: string]: BlockDefinitionInterface }
  userBlocks?: Array<{ id: string; name: string; content: string }>
  onUserBlocksUpdate: (blocks: EmailTemplateBlock[]) => Promise<void>
  urlParams?: UrlParams
  fileManagerSettings?: Partial<FilesSettings>
  onUpdateFileManagerSettings?: (settings: FilesSettings) => Promise<void>
  height?: number
}

// Create a default file manager settings that satisfies the FilesSettings interface
const defaultFileManagerSettings: FilesSettings = {
  endpoint: '',
  access_key: '',
  secret_key: '',
  bucket: '',
  region: '',
  cdn_endpoint: ''
}

export const DefaultEditor = (props: DefaultEditorProps) => {
  const {
    onChange,
    initialValue,
    defaultTemplateData: defaultTemplateData,
    blockDefinitions: customBlockDefinitions,
    userBlocks = [],
    onUserBlocksUpdate,
    urlParams = {},
    fileManagerSettings = {},
    onUpdateFileManagerSettings = async () => {},
    height = 700
  } = props

  const [form] = Form.useForm()
  const [tree, setTree] = useState<BlockInterface>(initialValue || createDefaultBlocks())
  const [containerHeight, setContainerHeight] = useState<number>(height)
  const containerRef = useRef<HTMLDivElement>(null)
  const [templateData, setTemplateData] = useState(
    defaultTemplateData ||
      `{
  "user": {
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com"
  },
  "double_opt_in_link": "https://example.com/double-opt-in",
  "unsubscribe_link": "https://example.com/unsubscribe",
  "open_tracking_pixel_src": "https://example.com/tracking.gif"
}`
  )

  // Combine default block definitions with any custom ones
  const blockDefinitions = {
    root: RootBlockDefinition,
    column: ColumnBlockDefinition,
    oneColumn: OneColumnBlockDefinition,
    columns168: Columns168BlockDefinition,
    columns204: Columns204BlockDefinition,
    columns420: Columns420BlockDefinition,
    columns816: Columns816BlockDefinition,
    columns888: Columns888BlockDefinition,
    columns1212: Columns1212BlockDefinition,
    columns6666: Columns6666BlockDefinition,
    image: ImageBlockDefinition,
    divider: DividerBlockDefinition,
    openTracking: OpenTrackingBlockDefinition,
    button: ButtonBlockDefinition,
    text: TextBlockDefinition,
    heading: HeadingBlockDefinition,
    liquid: LiquidTemplateBlockDefinition,
    ...customBlockDefinitions
  }

  // Merge provided settings with defaults to ensure all required properties exist
  const mergedFileManagerSettings: FilesSettings = {
    ...defaultFileManagerSettings,
    ...fileManagerSettings
  }

  // Update container height when it changes
  useEffect(() => {
    const updateHeight = () => {
      if (containerRef.current) {
        const height = containerRef.current.clientHeight || 700
        setContainerHeight(height)
      }
    }

    updateHeight()

    // Listen for resize events
    window.addEventListener('resize', updateHeight)

    return () => {
      window.removeEventListener('resize', updateHeight)
    }
  }, [])

  // Generate HTML and call onChange if provided
  const handleTreeChange = useCallback(
    (newTree: BlockInterface): void => {
      setTree(newTree)

      if (onChange) {
        const result: HtmlExportResult = ExportHTML(newTree, urlParams)
        if (!result.errors || result.errors.length === 0) {
          onChange(result.html)
        }
      }
    },
    [onChange, urlParams]
  )

  // Handler for test data updates
  const handleTemplateDataUpdate = useCallback(
    async (newData: string): Promise<void> => {
      setTemplateData(newData)
      form.setFieldsValue({ template_data: newData })
    },
    [form]
  )

  // Define container styles using StrictCSSProperties
  const containerStyles: StrictCSSProperties = {
    height: '100%',
    position: 'relative'
  }

  const emailEditorContainerStyles: StrictCSSProperties = {
    height: '100%',
    minHeight: '700px'
  }

  return (
    <div style={containerStyles}>
      <div ref={containerRef} className="email-editor-container" style={emailEditorContainerStyles}>
        <Form form={form} layout="vertical" initialValues={{ template_data: templateData }}>
          <Editor
            blockDefinitions={blockDefinitions}
            userBlocks={userBlocks}
            onUserBlocksUpdate={onUserBlocksUpdate}
            templateDataValue={templateData}
            selectedBlockId={'root'}
            value={tree}
            onChange={handleTreeChange}
            renderSelectedBlockButtons={(props: SelectedBlockButtonsProp) => (
              <SelectedBlockButtons {...props} />
            )}
            deviceWidth={DesktopWidth}
            urlParams={urlParams}
            fileManagerSettings={mergedFileManagerSettings}
            onUpdateFileManagerSettings={onUpdateFileManagerSettings}
            onUpdateTemplateData={handleTemplateDataUpdate}
          >
            <Layout form={form} macros={[]} height={containerHeight} />
          </Editor>
        </Form>
      </div>
    </div>
  )
}
