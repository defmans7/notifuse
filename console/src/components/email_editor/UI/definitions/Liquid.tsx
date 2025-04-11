import { Form, Space, Button, Modal, Alert } from 'antd'
import { Code } from 'lucide-react'
import { useRef, useState, useEffect } from 'react'
import { BlockDefinitionInterface, BlockRenderSettingsProps } from '../../Block'
import { BlockEditorRendererProps } from '../../BlockEditorRenderer'
import mjml2html from 'mjml-browser'
import AceInput from '../Widgets/AceInput'
import { usePrismjs } from '../Widgets/PrismJS'
import { Liquid } from 'liquidjs'
import { cloneDeep } from 'lodash'
import { returnUpdatedTree } from '../../Editor'
import { treeToMjml } from '../../utils'

const LiquidTemplateBlockDefinition: BlockDefinitionInterface = {
  name: 'Liquid + MJML',
  kind: 'liquid',
  containsDraggables: false,
  isDraggable: true,
  draggableIntoGroup: 'column',
  isDeletable: true,
  defaultData: {
    liquidCode: `{% if contact %}
<mj-text font-size="20px" color="#333333" font-family="helvetica">
  Hello {{ contact.first_name }}!
</mj-text>
<mj-text font-size="16px" color="#666666" font-family="helvetica">
  Email: {{ contact.email }}<br/>
  Phone: {{ contact.phone }}<br/>
  Country: {{ contact.country }}
</mj-text>
<mj-button background-color="#4CAF50" href="mailto:{{ contact.email }}">
  Contact Now
</mj-button>
{% else %}
<mj-text font-size="20px" color="#333333" font-family="helvetica">
  No Contact Provided
</mj-text>
<mj-text font-size="16px" color="#666666" font-family="helvetica">
  Please provide a contact to view their details.
</mj-text>
{% endif %}`
  },
  menuSettings: {},

  RenderSettings: (props: BlockRenderSettingsProps) => {
    const [liquidCode, setLiquidCode] = useState(props.block.data.liquidCode)
    const [liquidModalVisible, setLiquidModalVisible] = useState(false)
    const [modalHeight, setModalHeight] = useState(400)
    const [validationError, setValidationError] = useState<string | null>(null)
    const [templateDataModalVisible, setTemplateDataModalVisible] = useState(false)
    const [editingTemplateData, setEditingTemplateData] = useState(props.templateData)

    const handleUpdateTemplateData = async () => {
      try {
        await props.onUpdateTemplateData(editingTemplateData)
        setTemplateDataModalVisible(false)
      } catch (error) {
        console.error('Error updating test data:', error)
      }
    }

    useEffect(() => {
      const handleResize = () => {
        // Adjust modal height based on window height
        const windowHeight = window.innerHeight
        setModalHeight(Math.min(windowHeight * 0.7, 600))
      }

      window.addEventListener('resize', handleResize)
      handleResize() // Initialize on mount

      return () => {
        window.removeEventListener('resize', handleResize)
      }
    }, [])

    const validateCode = (code: string): string | null => {
      try {
        // First validate Liquid syntax
        const engine = new Liquid()
        engine.parse(code)

        const newBlock = cloneDeep(props.block)
        newBlock.data.liquidCode = code
        const newTree = returnUpdatedTree(props.tree, props.block.path, newBlock)

        const mjml = mjml2html(
          treeToMjml(newTree.data.styles, newTree, props.templateData, props.urlParams, undefined)
        )

        if (mjml.errors.length > 0) {
          return `MJML syntax error: ${mjml.errors[0].message}`
        }
        return null
      } catch (e: any) {
        return `Liquid syntax error: ${e.message}`
      }
    }

    const handleSave = () => {
      const error = validateCode(liquidCode)
      if (error) {
        setValidationError(error)
        return
      }

      setValidationError(null)
      props.block.data.liquidCode = liquidCode
      props.updateTree(props.block.path, props.block)
      setLiquidModalVisible(false)
    }

    // Format JSON for display
    const formattedTemplateData = (() => {
      try {
        return JSON.stringify(JSON.parse(props.templateData), null, 2)
      } catch (e) {
        return props.templateData
      }
    })()

    // Apply syntax highlighting using usePrismjs hook
    const templateDataRef = useRef<HTMLPreElement>(null)
    usePrismjs(templateDataRef)

    return (
      <>
        <div className="xpeditor-padding-h-l">
          <Form.Item
            label="Liquid Code"
            labelAlign="left"
            className="xpeditor-form-item-align-right"
            labelCol={{ span: 10 }}
            wrapperCol={{ span: 14 }}
            help="Liquid + MJML"
          >
            <Button type="primary" size="small" block onClick={() => setLiquidModalVisible(true)}>
              Edit Code
            </Button>
          </Form.Item>
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
              onClick={() => setTemplateDataModalVisible(true)}
            >
              Edit Test data
            </Button>
          </Form.Item>
        </div>
        <div style={{ height: `600px` }}>
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
          open={templateDataModalVisible}
          onOk={handleUpdateTemplateData}
          onCancel={() => setTemplateDataModalVisible(false)}
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

        <Modal
          title="Edit Liquid Code"
          open={liquidModalVisible}
          onCancel={() => {
            setLiquidModalVisible(false)
            setValidationError(null)
          }}
          width="80%"
          style={{ top: 20 }}
          styles={{
            body: { padding: 0, height: `${modalHeight}px`, overflow: 'hidden' }
          }}
          footer={[
            <Button
              key="cancel"
              onClick={() => {
                setLiquidModalVisible(false)
                setValidationError(null)
              }}
            >
              Cancel
            </Button>,
            <Button key="save" type="primary" onClick={handleSave}>
              Save
            </Button>
          ]}
        >
          {validationError && (
            <Alert
              message="Validation Error"
              description={validationError}
              type="error"
              style={{ margin: '8px' }}
            />
          )}
          <AceInput
            id="liquid-editor"
            mode="liquid"
            value={liquidCode}
            onChange={(value) => {
              setLiquidCode(value)
              setValidationError(null)
            }}
            height={`${modalHeight - 53}px`} // Adjust for Modal header height
            width="100%"
            theme="monokai"
          />
        </Modal>
      </>
    )
  },

  renderEditor: (props: BlockEditorRendererProps) => {
    const wrapperStyles: any = {
      position: 'relative'
    }

    const liquidCodeRef = useRef<HTMLPreElement>(null)

    try {
      // Use the usePrismjs hook instead of directly calling Prism.highlightElement
      // This will apply syntax highlighting to the code block
      usePrismjs(liquidCodeRef)
    } catch (error) {
      // Log the error but continue rendering
      console.error('Error applying Prism highlighting to Liquid code:', error)
    }

    // For preview mode, the liquid code will be parsed and executed in the Preview component
    // Here we just render the code with syntax highlighting
    return (
      <div style={wrapperStyles}>
        <div className="liquid-template-wrapper">
          <pre
            className="language-liquid"
            ref={liquidCodeRef}
            style={{ margin: 0, fontSize: '10px' }}
          >
            <code className="language-liquid">{props.block.data.liquidCode}</code>
          </pre>
        </div>
      </div>
    )
  },

  renderMenu: (_blockDefinition: BlockDefinitionInterface) => {
    return (
      <div className="xpeditor-ui-block">
        <Space size="middle">
          <Code size={16} style={{ marginTop: '5px' }} />
          Liquid + MJML
        </Space>
      </div>
    )
  }
}

export default LiquidTemplateBlockDefinition
