import { useState, useEffect, useRef } from 'react'
import { Button, Tooltip, Space, Form, Select, FormInstance, Modal } from 'antd'
import {
  DesktopOutlined,
  MobileOutlined,
  EyeOutlined,
  LeftOutlined,
  RightOutlined
} from '@ant-design/icons'
import _ from 'lodash'

import { BlockInterface } from '../Block'
import { useEditorContext, EditorContextValue } from '../Editor'
import { Blocks, BlocksProps } from './Blocks'
import Preview from './Preview'
import Settings from './Settings'
import AceInput from './Widgets/AceInput'
import { StrictCSSProperties, TemplateDataFormField } from '../types'
import { usePrismjs } from './Widgets/PrismJS'
import SimpleBar from 'simplebar-react'

export const MobileWidth = 400
export const DesktopWidth = 960

const FindBlockById = (currentBlock: BlockInterface, id: string): BlockInterface | undefined => {
  if (currentBlock.id === id) return currentBlock
  else if (currentBlock.children) {
    let found: BlockInterface | undefined
    currentBlock.children.forEach((child) => {
      const got = FindBlockById(child, id)
      if (got) found = got
    })
    return found
  }
  return undefined
}

interface LayoutProps {
  form: FormInstance<TemplateDataFormField>
  macros: Array<{ id: string; name: string }>
  height?: number
}

export const Layout = (props: LayoutProps) => {
  const editor: EditorContextValue = useEditorContext()
  const [isPreview, setIsPreview] = useState(false)
  const [layoutRightHeight, setLayoutRightHeight] = useState(400)
  const mainRef = useRef<HTMLDivElement>(null)
  const rightPanelRef = useRef<HTMLDivElement>(null)
  const templateDataTitleRef = useRef<HTMLDivElement>(null)
  const macroSelectRef = useRef<HTMLDivElement>(null)
  const templateDataRef = useRef<HTMLPreElement>(null)

  // State for the edit modal
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [editingTemplateData, setEditingTemplateData] = useState('')

  // Get the current test data value from the form
  const templateData = Form.useWatch('template_data', props.form) || editor.templateDataValue

  // Format JSON for display
  const formattedTemplateData = (() => {
    try {
      return JSON.stringify(JSON.parse(templateData), null, 2)
    } catch (e) {
      return templateData
    }
  })()

  // Apply syntax highlighting using usePrismjs hook
  usePrismjs(templateDataRef)

  // Handle updating test data
  const handleUpdateTemplateData = async () => {
    try {
      // Format the JSON
      // const formatted = JSON.stringify(JSON.parse(editingTemplateData), null, 2)
      // Update the editor context
      await editor.onUpdateTemplateData(editingTemplateData)
      // Close the modal
      setIsModalVisible(false)
    } catch (e) {
      // Show error notification
      console.error('Invalid JSON:', e)
      // Keep the modal open so user can fix the JSON
    }
  }

  useEffect(() => {
    const updateHeight = () => {
      if (mainRef.current && rightPanelRef.current) {
        // Get the total height of the right panel
        const rightPanelHeight = rightPanelRef.current.clientHeight

        // Calculate the height taken by other elements
        let otherElementsHeight = 0

        if (templateDataTitleRef.current) {
          otherElementsHeight += templateDataTitleRef.current.clientHeight
        }

        if (macroSelectRef.current) {
          otherElementsHeight += macroSelectRef.current.clientHeight
        }

        // Add padding and margins - adjust these values as needed
        const padding = 25

        // Calculate the remaining height for the AceInput
        const availableHeight = rightPanelHeight - otherElementsHeight - padding

        setLayoutRightHeight(availableHeight > 200 ? availableHeight : 400)
      }
    }

    updateHeight()

    // Add resize listener
    window.addEventListener('resize', updateHeight)

    // Also run when preview mode changes or macro selection changes
    if (isPreview) {
      setTimeout(updateHeight, 100)
    }

    // Clean up
    return () => {
      window.removeEventListener('resize', updateHeight)
    }
  }, [isPreview, props.macros])

  // console.log('render')

  if (editor.currentTree.kind !== 'root') {
    return <>First block should be "root", got: {editor.currentTree.kind}</>
  }

  const blocksProps: BlocksProps = {
    blockDefinitions: editor.blockDefinitions,
    userBlocks: editor.userBlocks,
    onUserBlocksUpdate: editor.onUserBlocksUpdate,
    renderBlockForMenu: editor.renderBlockForMenu,
    renderSavedBlockForMenu: editor.renderSavedBlockForMenu
  }

  const pathBlocks: BlockInterface[] = [editor.currentTree]

  let currentPath = ''

  let selectedBlock = FindBlockById(editor.currentTree, editor.selectedBlockId)

  // focus root by default
  if (!selectedBlock) {
    selectedBlock = editor.currentTree
  }

  selectedBlock.path.split('.').forEach((part) => {
    if (currentPath === '') {
      currentPath = part
    } else {
      currentPath += '.' + part
    }

    const block: BlockInterface = _.get(editor.currentTree, currentPath)

    if (block && block.kind) {
      pathBlocks.push(block)
    }
  })

  const togglePreview = (): void => {
    setIsPreview(!isPreview)
  }

  const toggleDevice = (): void => {
    if (editor.deviceWidth === MobileWidth) {
      editor.setDeviceWidth(DesktopWidth)
    } else {
      editor.setDeviceWidth(MobileWidth)
    }
  }

  const goBackHistory = (): void => {
    const lastHistoryIndex: number = editor.history.length - 1
    if (lastHistoryIndex > 0) {
      // console.log('back to', editor.currentHistoryIndex - 1)
      editor.setCurrentHistoryIndex(editor.currentHistoryIndex - 1)
    }
  }

  const goNextHistory = (): void => {
    const lastHistoryIndex: number = editor.history.length - 1
    if (editor.currentHistoryIndex < lastHistoryIndex) {
      editor.setCurrentHistoryIndex(editor.currentHistoryIndex + 1)
    }
  }

  // console.log('layout props', props)

  const mainStyle: StrictCSSProperties = {
    height: props.height || '100vh'
  }

  return (
    <div className="xpeditor-main" style={mainStyle} ref={mainRef}>
      <div className={'xpeditor-layout-left'}>
        {!isPreview && (
          <SimpleBar style={{ maxHeight: '100%' }}>
            <Blocks {...blocksProps} />
          </SimpleBar>
        )}
      </div>

      <div className="xpeditor-layout-middle">
        {isPreview && (
          <>
            <Preview
              tree={editor.currentTree}
              templateData={templateData}
              isMobile={editor.deviceWidth === MobileWidth}
              deviceWidth={editor.deviceWidth}
              toggleDevice={toggleDevice}
              urlParams={editor.urlParams}
              closePreview={togglePreview}
            />
          </>
        )}

        {!isPreview && (
          <>
            <div className="xpeditor-topbar">
              <span style={{ float: 'right' }}>
                <Space>
                  <Space.Compact>
                    <Button
                      size="small"
                      type="text"
                      disabled={editor.deviceWidth === MobileWidth}
                      onClick={() => toggleDevice()}
                    >
                      <MobileOutlined />
                    </Button>
                    <Button
                      size="small"
                      type="text"
                      disabled={editor.deviceWidth === DesktopWidth}
                      onClick={() => toggleDevice()}
                    >
                      <DesktopOutlined />
                    </Button>
                  </Space.Compact>

                  <Button type="primary" size="small" ghost onClick={() => togglePreview()}>
                    <EyeOutlined />
                    &nbsp; Preview
                  </Button>
                </Space>
              </span>

              <Space size="large">
                <>
                  <Space.Compact>
                    <Tooltip title="Undo">
                      <Button
                        size="small"
                        type="text"
                        onClick={goBackHistory}
                        disabled={editor.currentHistoryIndex === 0}
                        icon={<LeftOutlined />}
                      />
                    </Tooltip>
                    <Tooltip title="Redo">
                      <Button
                        size="small"
                        type="text"
                        onClick={goNextHistory}
                        disabled={editor.currentHistoryIndex === editor.history.length - 1}
                        icon={<RightOutlined />}
                      />
                    </Tooltip>
                  </Space.Compact>
                  <div className="xpeditor-path">
                    {pathBlocks.map((block, i) => {
                      const isLast = i === pathBlocks.length - 1 ? true : false
                      return (
                        <span key={i}>
                          {isLast === true && (
                            <span className="xpeditor-path-item-last">
                              {editor.blockDefinitions[block.kind]?.name}
                            </span>
                          )}
                          {isLast === false && (
                            <>
                              <span
                                className="xpeditor-path-item"
                                onClick={editor.selectBlock.bind(null, block)}
                              >
                                {editor.blockDefinitions[block.kind]?.name}
                              </span>
                              <span className="xpeditor-path-divider">/</span>
                            </>
                          )}
                        </span>
                      )
                    })}
                  </div>
                </>
              </Space>
            </div>
            <div onClick={editor.selectBlock.bind(null, editor.currentTree)}>{editor.editor}</div>
          </>
        )}
      </div>

      <div className={'xpeditor-layout-right'} ref={rightPanelRef}>
        {!isPreview && (
          <Settings
            block={selectedBlock}
            blockDefinition={editor.blockDefinitions[selectedBlock.kind]}
            tree={editor.currentTree}
            updateTree={editor.updateTree}
            urlParams={editor.urlParams}
            templateData={editor.templateDataValue}
            onUpdateTemplateData={editor.onUpdateTemplateData}
          />
        )}
        {isPreview && (
          <>
            <div className="xpeditor-ui-menu-title" ref={templateDataTitleRef}>
              Preview Settings
            </div>
            <div ref={macroSelectRef}>
              {props.macros && props.macros.length > 0 && (
                <Form.Item
                  label="Use a macros page"
                  name="template_macro_id"
                  style={{ padding: '0 12px' }}
                >
                  <Select
                    style={{ width: '100%' }}
                    popupMatchSelectWidth={false}
                    allowClear={true}
                    size="small"
                    placeholder="Select macros page"
                    options={props.macros.map((x) => {
                      return { label: x.name, value: x.id }
                    })}
                    onChange={(val: string) =>
                      props.form.setFieldsValue({ template_macro_id: val })
                    }
                  />
                </Form.Item>
              )}
            </div>

            <div className="xpeditor-padding-h-l" style={{ paddingTop: '12px' }}>
              <Form.Item
                label="Test data"
                labelAlign="left"
                className="xpeditor-form-item-align-right"
                labelCol={{ span: 10 }}
                wrapperCol={{ span: 14 }}
              >
                <Button
                  type="primary"
                  size="small"
                  block
                  onClick={() => {
                    setEditingTemplateData(templateData)
                    setIsModalVisible(true)
                  }}
                >
                  Edit Test data
                </Button>
              </Form.Item>
            </div>
            <div style={{ height: `${layoutRightHeight}px` }}>
              <pre
                className="language-json"
                ref={templateDataRef}
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
                <code className="language-json">{formattedTemplateData}</code>
              </pre>
            </div>

            <Modal
              title="Edit Test Data"
              open={isModalVisible}
              onOk={handleUpdateTemplateData}
              onCancel={() => setIsModalVisible(false)}
              width="80%"
              style={{ top: 20 }}
              styles={{
                body: {
                  padding: 0,
                  height: '70vh',
                  overflow: 'hidden'
                }
              }}
            >
              <AceInput
                id="test_data_editor"
                mode="json"
                value={editingTemplateData}
                onChange={(val: string) => setEditingTemplateData(val)}
                height="calc(70vh - 53px)" // Adjust for Modal header height
                width="100%"
                theme="monokai"
              />
            </Modal>
          </>
        )}
      </div>
    </div>
  )
}
