import { ReactNode, useMemo } from 'react'
import { BlockDefinitionMap, BlockDefinitionInterface, BlockInterface } from '../Block'
import { Collapse, Tooltip, Popconfirm } from 'antd'
import { truncate } from 'lodash'
import { EmailTemplateBlock } from '../types'
import { DeleteOutlined } from '@ant-design/icons'

export interface BlocksProps {
  blockDefinitions: BlockDefinitionMap
  userBlocks: EmailTemplateBlock[]
  onUserBlocksUpdate: (blocks: EmailTemplateBlock[]) => Promise<void>
  renderBlockForMenu: (blockDefinition: BlockDefinitionInterface) => ReactNode
  renderSavedBlockForMenu: (block: BlockInterface, renderMenu: ReactNode) => ReactNode
}

export const Blocks = (props: BlocksProps) => {
  const handleDeleteBlock = async (blockId: string) => {
    const updatedBlocks = props.userBlocks.filter((block) => block.id !== blockId)
    await props.onUserBlocksUpdate(updatedBlocks)
  }

  const contentItems = useMemo(() => {
    return {
      key: '1',
      label: (
        <span className="xpeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>
          Content blocks
        </span>
      ),
      style: { paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 },
      children: (
        <>
          {props.renderBlockForMenu(props.blockDefinitions['image'])}
          {props.renderBlockForMenu(props.blockDefinitions['button'])}
          {props.renderBlockForMenu(props.blockDefinitions['heading'])}
          {props.renderBlockForMenu(props.blockDefinitions['text'])}
          {props.renderBlockForMenu(props.blockDefinitions['liquid'])}
          {props.renderBlockForMenu(props.blockDefinitions['divider'])}
          {props.renderBlockForMenu(props.blockDefinitions['openTracking'])}
        </>
      )
    }
  }, [props.blockDefinitions, props.renderBlockForMenu])

  const savedItems = useMemo(() => {
    if (!props.userBlocks || props.userBlocks.length === 0) {
      return null
    }

    return {
      key: '2',
      label: (
        <span className="xpeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>
          Saved blocks ({props.userBlocks.length})
        </span>
      ),
      style: { paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 },
      children: (
        <>
          {props.userBlocks.map((b: EmailTemplateBlock, i: number) => {
            return (
              <div key={i}>
                {props.renderSavedBlockForMenu(
                  JSON.parse(b.content),
                  <Tooltip title={b.name}>
                    <div className="xpeditor-ui-saved-block">
                      <Popconfirm
                        title="Are you sure to delete this block?"
                        onConfirm={() => handleDeleteBlock(b.id)}
                        okText="Yes"
                        cancelText="No"
                      >
                        <div
                          className="xpeditor-ui-saved-block-delete"
                          style={{ cursor: 'pointer' }}
                        >
                          <DeleteOutlined />
                        </div>
                      </Popconfirm>
                      {truncate(b.name, { length: 20 })}
                    </div>
                  </Tooltip>
                )}
              </div>
            )
          })}
        </>
      )
    }
  }, [props.userBlocks, props.renderSavedBlockForMenu, handleDeleteBlock])

  const layoutItems = useMemo(() => {
    return {
      key: '3',
      label: (
        <span className="xpeditor-ui-menu-title" style={{ padding: 0, margin: 0 }}>
          Layout sections
        </span>
      ),
      style: { paddingLeft: 0, paddingRight: 0, paddingTop: 12, marginBottom: 0 },
      children: (
        <>
          {props.renderBlockForMenu(props.blockDefinitions['oneColumn'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns1212'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns888'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns6666'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns420'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns816'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns168'])}
          {props.renderBlockForMenu(props.blockDefinitions['columns204'])}
        </>
      )
    }
  }, [props.blockDefinitions, props.renderBlockForMenu])

  const collapseItems = useMemo(() => {
    const items = [contentItems]

    if (savedItems) {
      items.push(savedItems)
    }

    items.push(layoutItems)

    return items
  }, [contentItems, savedItems, layoutItems])

  return (
    <>
      <Collapse defaultActiveKey={['1', '3']} ghost items={collapseItems} />
    </>
  )
}
