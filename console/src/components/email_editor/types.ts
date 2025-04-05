/**
 * Shared TypeScript types for the email editor
 */

import { CSSProperties, ReactNode, RefObject } from 'react'
import { BlockInterface, BlockDefinitionMap } from './Block'

/**
 * URL parameter record type for tracking links
 */
export interface UrlParams {
  utm_source?: string
  utm_medium?: string
  utm_campaign?: string
  utm_content?: string
  utm_term?: string
  utm_id?: string
  [key: string]: string | undefined
}

/**
 * DOM node reference type
 */
export type DomNodeRef = RefObject<HTMLDivElement | null> | undefined

/**
 * Focus event handler type
 */
export type FocusEventHandler = (node: DomNodeRef) => void

/**
 * React CSS styles with type safety
 */
export interface StrictCSSProperties extends CSSProperties {
  position?: 'relative' | 'absolute' | 'fixed' | 'sticky' | 'static'
  display?: 'block' | 'inline' | 'flex' | 'inline-flex' | 'table' | 'table-cell' | 'none'
  textAlign?: 'left' | 'center' | 'right' | 'justify'
  float?: 'left' | 'right' | 'none'
  overflow?: 'hidden' | 'auto' | 'scroll' | 'visible'
  fontWeight?: 'normal' | 'bold' | 'lighter' | 'bolder' | number
}

/**
 * Block update handler type
 */
export type BlockUpdateHandler = (path: string, data: unknown) => void

/**
 * Rendered block buttons props
 */
export interface BlockButtonsProps {
  isDraggable: boolean
  blockDefinitions: BlockDefinitionMap
  block: BlockInterface
  deleteBlock: (block: BlockInterface) => void
  cloneBlock: (block: BlockInterface) => void
  existingBlocks: EmailTemplateBlock[]
  onExistingBlocksUpdate: (blocks: EmailTemplateBlock[]) => Promise<void>
}

/**
 * Email template block type
 */
export interface EmailTemplateBlock {
  id: string
  name: string
  content: string // json tree
}

/**
 * Block drop result from smooth-dnd
 */
export interface DropResultPayload {
  removedIndex: number | null
  addedIndex: number | null
  payload: BlockInterface
}

/**
 * Form field types for test data
 */
export interface TemplateDataFormField {
  template_data: string
  template_macro_id?: string
}

/**
 * Drag and drop container props
 */
export interface DragContainerProps {
  groupName?: string
  behaviour?: 'move' | 'copy'
  getChildPayload: (index: number) => BlockInterface | (() => BlockInterface)
  children?: ReactNode
  style?: StrictCSSProperties
  dragHandleSelector?: string
  dragClass?: string
  dropClass?: string
  dropPlaceholder?: {
    animationDuration: number
    showOnTop: boolean
    className: string
  }
  onDrop?: (result: DropResultPayload) => void
}

/**
 * HTML export result
 */
export interface HtmlExportResult {
  html: string
  errors?: any[] // Using any[] for now to accommodate both string[] and MJMLParseError[]
}
