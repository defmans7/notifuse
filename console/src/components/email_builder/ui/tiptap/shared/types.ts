import React from 'react'

// Base props shared by all Tiptap editor variants
export interface BaseTiptapProps {
  content?: string
  onChange?: (content: string) => void
  readOnly?: boolean
  placeholder?: string
  containerStyle?: React.CSSProperties
  autoFocus?: boolean
  buttons?: string[] // Array of button names to display
}

// Props specific to the rich text editor
export interface TiptapRichEditorProps extends BaseTiptapProps {
  // Additional props specific to rich editor can be added here
}

// Props specific to the inline editor
export interface TiptapInlineEditorProps extends BaseTiptapProps {
  // Additional props specific to inline editor can be added here
}

// Toolbar props
export interface TiptapToolbarProps {
  editor: any
  buttons?: string[]
  mode?: 'rich' | 'inline'
}

// Individual toolbar button props
export interface ToolbarButtonProps {
  onClick?: () => void
  disabled?: boolean
  isActive?: boolean
  title: string
  children: React.ReactNode
}

// Color button props
export interface ColorButtonProps {
  icon: any
  currentColor?: string
  onColorChange: (color: string) => void
  title: string
  isActive?: boolean
}

// Emoji button props
export interface EmojiButtonProps {
  onEmojiSelect: (emoji: any) => void
  title: string
}

// Link button props
export interface LinkButtonProps {
  editor: any
  title: string
}

// Available button types
export type ButtonType =
  | 'undo'
  | 'redo'
  | 'bold'
  | 'italic'
  | 'underline'
  | 'strikethrough'
  | 'textColor'
  | 'backgroundColor'
  | 'emoji'
  | 'link'
  | 'superscript'
  | 'subscript'

// Link types
export type LinkType = 'url' | 'email' | 'phone' | 'anchor'
