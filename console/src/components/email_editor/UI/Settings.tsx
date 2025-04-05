import { BlockInterface, BlockDefinitionInterface, BlockRenderSettingsProps } from '../Block'
import cloneDeep from 'lodash/cloneDeep'
import SimpleBar from 'simplebar-react'
import get from 'lodash/get'

interface SettingsProps {
  block: BlockInterface
  blockDefinition: BlockDefinitionInterface
  updateTree: (path: string, data: any) => void
  tree: BlockInterface
  urlParams: any
  templateData: any
  onUpdateTemplateData: (templateData: any) => Promise<void>
}

const Settings = (props: SettingsProps) => {
  // const updateSettings = (settings: any) => {
  //     // console.log('new settings are', settings.styles.backgroundColor)
  //     const newBlock = cloneDeep(props.block)
  //     newblock.data = settings
  //     props.updateTree(newBlock.path, newBlock)
  // }

  if (!props.blockDefinition) {
    return <div>Block definition {props.block.kind} not found.</div>
  }

  // avoid to mutate the block
  // const safeBlock = cloneDeep(props.block)
  const safeTree = cloneDeep(props.tree)
  const block = props.block.path === '' ? safeTree : get(safeTree, props.block.path)

  const settingsProps: BlockRenderSettingsProps = {
    block: block,
    updateTree: props.updateTree,
    tree: safeTree,
    urlParams: props.urlParams,
    templateData: props.templateData,
    onUpdateTemplateData: props.onUpdateTemplateData
  }

  const RenderSettingsComponent = props.blockDefinition.RenderSettings

  return (
    <SimpleBar style={{ maxHeight: '100%' }}>
      <div className="xpeditor-ui-menu-title">Current block settings</div>
      <RenderSettingsComponent {...settingsProps} />
    </SimpleBar>
  )
}

export default Settings
