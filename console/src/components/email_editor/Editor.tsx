import {
  useState,
  MouseEvent,
  ReactNode,
  useEffect,
  createContext,
  useContext,
  useRef,
  useCallback
} from 'react'
import { cloneDeep, get, remove, set } from 'lodash'
import uuid from 'short-uuid'
import { BlockEditorRenderer } from './BlockEditorRenderer'
import { BlockDefinitionInterface, BlockInterface, BlockDefinitionMap } from './Block'
import Container from './Container'
import Draggable from './Draggable'
import { DropResult } from './smooth-dnd'
import {
  UrlParams,
  DomNodeRef,
  BlockUpdateHandler,
  BlockButtonsProps,
  EmailTemplateBlock
} from './types'
import './UI/editor.css'
import { FileManagerSettings } from '../../services/api/types'

const EditorContext = createContext<EditorContextValue | null>(null)

export function useEditorContext(): EditorContextValue {
  const editorValue = useContext(EditorContext)
  if (!editorValue) {
    throw new Error('Missing EditorContextProvider in its parent.')
  }
  return editorValue
}

export interface EditorContextValue {
  blockDefinitions: BlockDefinitionMap
  userBlocks: EmailTemplateBlock[]
  onUserBlocksUpdate: (blocks: EmailTemplateBlock[]) => Promise<void>
  templateDataValue: string
  currentTree: BlockInterface
  selectedBlockId: string
  updateTree: BlockUpdateHandler
  selectBlock: (block: BlockInterface, event?: MouseEvent) => void
  renderBlockForMenu: (blockDefinition: BlockDefinitionInterface) => ReactNode
  renderSavedBlockForMenu: (block: BlockInterface, renderMenu: ReactNode) => ReactNode
  editor: ReactNode
  history: BlockInterface[]
  currentHistoryIndex: number
  setCurrentHistoryIndex: (index: number) => void
  deviceWidth: number
  setDeviceWidth: (width: number) => void
  urlParams: UrlParams
  onFocusBlock: (node: DomNodeRef) => void
  fileManagerSettings?: FileManagerSettings
  onUpdateFileManagerSettings: (settings: FileManagerSettings) => Promise<void>
  onUpdateTemplateData: (templateData: string) => Promise<void>
}

export type SelectedBlockButtonsProp = BlockButtonsProps

export interface EditorProps {
  children: ReactNode
  blockDefinitions: BlockDefinitionMap
  templateDataValue: string
  value: BlockInterface
  onChange: (newValue: BlockInterface) => void
  renderSelectedBlockButtons: (props: SelectedBlockButtonsProp) => ReactNode
  deviceWidth: number
  selectedBlockId?: string
  urlParams: UrlParams
  userBlocks: EmailTemplateBlock[]
  onUserBlocksUpdate: (blocks: EmailTemplateBlock[]) => Promise<void>
  fileManagerSettings?: FileManagerSettings
  onUpdateFileManagerSettings: (settings: FileManagerSettings) => Promise<void>
  onUpdateTemplateData?: (templateData: string) => Promise<void>
}

// recursive id generation, used to clone blocks
export const generateNewBlockIds = (block: BlockInterface): void => {
  block.id = uuid.generate()
  // generate new uuids for children
  if (block.children) {
    block.children.forEach((child) => {
      generateNewBlockIds(child)
    })
  }
}

export const returnUpdatedTree = (currentTree: BlockInterface, path: string, data: any) => {
  let newTree: BlockInterface

  if (path === '') {
    newTree = cloneDeep(data) as BlockInterface
  } else {
    newTree = cloneDeep(currentTree)
    set(newTree, path, data)
  }

  return newTree
}

export const Editor = (props: EditorProps) => {
  const focusedNodeRef = useRef<DomNodeRef>(undefined)

  useEffect(() => {
    return () => {
      // reset focused node on cleanup
      if (focusedNodeRef.current && focusedNodeRef.current.current) {
        focusedNodeRef.current.current.classList.remove('xpeditor-focused')
      }
    }
  }, [])

  const recomputeBlockpaths = useCallback((block: BlockInterface): void => {
    if (block.children) {
      block.children.forEach((child, i) => {
        child.path = block.path + '.children[' + i + ']'
        recomputeBlockpaths(child)
      })
    }
  }, [])

  // set initial block paths from provided tree
  recomputeBlockpaths(props.value)

  // history is a ref to avoid race conditions
  const historyRef = useRef<BlockInterface[]>([props.value])
  const [currentHistoryIndex, setCurrentHistoryIndex] = useState<number>(0)
  const [selectedBlockId, setSelectedBlockId] = useState<string>(
    props.selectedBlockId ? props.selectedBlockId : props.value.id
  )
  const [deviceWidth, setDeviceWidth] = useState<number>(props.deviceWidth)

  const onContainerDrop = useCallback(
    (newBlockPathInTree: string, dropResult: DropResult): void => {
      const { removedIndex, addedIndex } = dropResult as {
        removedIndex: number | null
        addedIndex: number | null
      }

      // abort if nothing
      if (removedIndex === null && addedIndex === null) return

      // get children at path
      const finalPath = newBlockPathInTree === '' ? 'children' : newBlockPathInTree + '.children'
      // console.log('finalPath:', finalPath)

      // Get the last tree from history
      const currentTree = historyRef.current[historyRef.current.length - 1]
      // Create a deep clone of the current tree to make all modifications on
      const newTree = cloneDeep(currentTree)

      // Get the destination container's children
      const destinationChildren = get(newTree, finalPath)
      if (!destinationChildren) return

      const movedBlock: BlockInterface = dropResult.payload as BlockInterface

      // console.log('Drop result:', newBlockPathInTree, dropResult)
      // console.log('Tree:', newTree)

      let itemToAdd = cloneDeep(movedBlock)

      // Case 1: Block is removed from same container
      if (removedIndex !== null) {
        itemToAdd = destinationChildren.splice(removedIndex, 1)[0]
      }

      // Check if payload is from another container (already has an ID and path)
      const isFromDifferentContainer =
        movedBlock &&
        movedBlock.path &&
        !movedBlock.path.startsWith(newBlockPathInTree) &&
        movedBlock.path !== ''

      if (isFromDifferentContainer) {
        // console.log('Block from different container:', movedBlock.path)

        // This is a block from another container, we need to:
        // 1. Remove it from its original location
        const sourcePath = movedBlock.path.substring(0, movedBlock.path.lastIndexOf('['))
        const sourceParentPath = sourcePath.substring(0, sourcePath.lastIndexOf('.'))
        // const sourceIndex = parseInt(
        //   movedBlock.path.substring(
        //     movedBlock.path.lastIndexOf('[') + 1,
        //     movedBlock.path.lastIndexOf(']')
        //   )
        // )
        // console.log('Source details:', { sourcePath, sourceParentPath, sourceIndex })

        // Get the parent's children array from our cloned tree
        const sourceParent = get(newTree, sourceParentPath)
        if (sourceParent && sourceParent.children) {
          // Find the actual block in the source container
          const sourceBlockIndex = sourceParent.children.findIndex(
            (child: BlockInterface) => child.id === movedBlock.id
          )

          if (sourceBlockIndex !== -1) {
            // Use the actual block from the source (not the payload which might be stale)
            itemToAdd = sourceParent.children[sourceBlockIndex]
            // Remove it from source
            sourceParent.children.splice(sourceBlockIndex, 1)
            // console.log('Removed block from source container')
          }
        }
      }

      // Add the item to the destination
      if (addedIndex !== null) {
        destinationChildren.splice(addedIndex, 0, itemToAdd)
        // console.log('Added block to destination container')
      }

      // Apply all the changes at once
      recomputeBlockpaths(newTree)
      // console.log('Recomputed paths for new tree', newTree)

      props.onChange(newTree)

      // append to history
      historyRef.current = [...historyRef.current, newTree]
      // move cursor to last version
      setCurrentHistoryIndex(historyRef.current.length - 1)

      // Select the moved block
      // console.log('Selecting block:', ite=mToAdd)
      selectBlock(movedBlock)
    },
    [currentHistoryIndex, recomputeBlockpaths, props.onChange]
  )

  const selectBlock = useCallback(
    (block: BlockInterface, event?: MouseEvent): void => {
      if (event) {
        event.preventDefault()
        event.stopPropagation()
      }

      if (selectedBlockId !== block.id) {
        setSelectedBlockId(block.id)
        if (block.kind === 'root') {
          onFocusBlock(undefined)
        }
      }
    },
    [selectedBlockId]
  )

  const updateTree = useCallback<BlockUpdateHandler>(
    (path, data) => {
      const currentTree = historyRef.current[currentHistoryIndex]
      const newTree = returnUpdatedTree(currentTree, path, data)

      props.onChange(newTree)

      // append to history
      historyRef.current = [...historyRef.current, newTree]
      // move cursor to last version
      setCurrentHistoryIndex(historyRef.current.length - 1)
    },
    [currentHistoryIndex, props.onChange]
  )

  const getParentBlock = useCallback(
    (tree: BlockInterface, block: BlockInterface): BlockInterface => {
      const parts = block.path.split('.')

      // get parent block path
      let parentBlock = tree
      let parentPath = ''

      // find parent block
      parts.forEach((part, i) => {
        // traverse tree as long as we dont reach the last block
        if (i < parts.length - 1) {
          parentBlock = get(tree, parentPath + (i === 0 ? '' : '.') + part)
          parentPath = parentBlock.path
        }
      })

      return parentBlock
    },
    []
  )

  const deleteBlock = useCallback(
    (block: BlockInterface): void => {
      if (!props.blockDefinitions[block.kind].isDeletable) {
        alert('The block ' + block.kind + ' is not deletable')
        return
      }

      const currentTree = historyRef.current[currentHistoryIndex]

      const newTree = cloneDeep(currentTree)

      const parentBlock = getParentBlock(newTree, block)

      parentBlock.children = remove(parentBlock.children, (child) => child.id !== block.id)

      recomputeBlockpaths(parentBlock)

      updateTree(
        parentBlock.path + (parentBlock.path === '' ? '' : '.') + 'children',
        parentBlock.children
      )
    },
    [props.blockDefinitions, currentHistoryIndex, getParentBlock, recomputeBlockpaths, updateTree]
  )

  const cloneBlock = useCallback(
    (block: BlockInterface): void => {
      const currentTree = historyRef.current[currentHistoryIndex]

      const newTree = cloneDeep(currentTree)

      const parentBlock = getParentBlock(newTree, block)

      if (!parentBlock.children) {
        parentBlock.children = []
      }

      const newBlock = cloneDeep(block)

      // append after block
      const currentBlockIndex = parentBlock.children.findIndex((child) => child.id === block.id)
      const newBlockIndex = currentBlockIndex + 1

      newBlock.path =
        parentBlock.path + (parentBlock.path === '' ? '' : '.') + 'children[' + newBlockIndex + ']'
      generateNewBlockIds(newBlock)

      parentBlock.children.splice(newBlockIndex, 0, newBlock)
      recomputeBlockpaths(parentBlock)

      updateTree(parentBlock.path, parentBlock)
    },
    [currentHistoryIndex, getParentBlock, recomputeBlockpaths, updateTree]
  )

  const generateBlockFromDefinition = useCallback(
    (blockDefinition: BlockDefinitionInterface): BlockInterface => {
      const id = uuid.generate()

      const block: BlockInterface = {
        id: id,
        kind: blockDefinition.kind,
        path: '', // path is set when rendering
        children: blockDefinition.children
          ? blockDefinition.children.map((child) => {
              return generateBlockFromDefinition(child)
            })
          : [],
        data: { ...blockDefinition.defaultData }
      }

      return block
    },
    []
  )

  const renderBlockForMenu = useCallback(
    (blockDefinition: BlockDefinitionInterface): ReactNode => {
      return (
        <Container
          key={blockDefinition.kind}
          groupName={blockDefinition.draggableIntoGroup}
          behaviour="copy"
          getChildPayload={generateBlockFromDefinition.bind(null, blockDefinition)}
        >
          <Draggable>
            {blockDefinition.renderMenu
              ? blockDefinition.renderMenu(blockDefinition)
              : 'renderMenu() not provided for: ' + blockDefinition.kind}
          </Draggable>
        </Container>
      )
    },
    [generateBlockFromDefinition]
  )

  const renderSavedBlockForMenu = useCallback(
    (block: BlockInterface, renderMenu: ReactNode): ReactNode => {
      // find definition of block
      if (!props.blockDefinitions[block.kind]) {
        console.error('block definition not found for block', block)
        return ''
      }

      return (
        <Container
          key={block.id}
          groupName={props.blockDefinitions[block.kind].draggableIntoGroup}
          behaviour="copy"
          getChildPayload={() => {
            generateNewBlockIds(block)
            return block
          }}
        >
          <Draggable>{renderMenu}</Draggable>
        </Container>
      )
    },
    [props.blockDefinitions]
  )

  const onFocusBlock = useCallback((node: DomNodeRef): void => {
    const previousNode = focusedNodeRef.current?.current
    const currentNode = node?.current

    // abort if the focus is on same block
    if (previousNode && currentNode && previousNode.id === currentNode.id) {
      return
    }

    // remove previous CSS if possible
    if (previousNode) {
      previousNode.classList.remove('xpeditor-focused')
    }

    if (currentNode) {
      currentNode.classList.add('xpeditor-focused')
    }

    focusedNodeRef.current = node
  }, [])

  if (props.value.kind !== 'root') {
    return <>First block should be "root", got: {props.value.kind}</>
  }

  const currentTree = historyRef.current[currentHistoryIndex]

  const blockEditorRendererProps = {
    block: currentTree,
    blockDefinitions: props.blockDefinitions,
    onContainerDrop: onContainerDrop,
    onSelectBlock: selectBlock,
    selectedBlockId: selectedBlockId,
    onFocusBlock: onFocusBlock,
    renderSelectedBlockButtons: (buttonProps: BlockButtonsProps) =>
      props.renderSelectedBlockButtons({
        ...buttonProps,
        existingBlocks: props.userBlocks,
        onExistingBlocksUpdate: props.onUserBlocksUpdate
      }),
    deleteBlock: deleteBlock,
    cloneBlock: cloneBlock,
    updateTree: updateTree,
    tree: props.value,
    deviceWidth: deviceWidth
  }

  const layoutProps: EditorContextValue = {
    blockDefinitions: props.blockDefinitions,
    userBlocks: props.userBlocks,
    onUserBlocksUpdate: props.onUserBlocksUpdate,
    templateDataValue: props.templateDataValue,
    currentTree: currentTree,
    selectedBlockId: selectedBlockId,
    updateTree: updateTree,
    selectBlock: selectBlock,
    renderBlockForMenu: renderBlockForMenu,
    renderSavedBlockForMenu: renderSavedBlockForMenu,
    editor: (
      <div
        onMouseLeave={() => {
          // remove focus when leaving the editor
          if (focusedNodeRef.current) {
            onFocusBlock(undefined)
          }
        }}
      >
        <BlockEditorRenderer {...blockEditorRendererProps} />
      </div>
    ),
    history: historyRef.current,
    currentHistoryIndex: currentHistoryIndex,
    setCurrentHistoryIndex: setCurrentHistoryIndex,
    deviceWidth: deviceWidth,
    setDeviceWidth: setDeviceWidth,
    urlParams: props.urlParams,
    onFocusBlock: onFocusBlock,
    fileManagerSettings: props.fileManagerSettings,
    onUpdateFileManagerSettings: props.onUpdateFileManagerSettings,
    onUpdateTemplateData: props.onUpdateTemplateData || (() => Promise.resolve())
  }

  return <EditorContext.Provider value={layoutProps}>{props.children}</EditorContext.Provider>
}
