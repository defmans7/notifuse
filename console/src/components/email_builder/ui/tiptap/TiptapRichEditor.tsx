import React, { useEffect, useCallback, useRef } from 'react'
import { useEditor, EditorContent } from '@tiptap/react'
import type { TiptapRichEditorProps } from './shared/types'
import { createRichExtensions } from './shared/extensions'
import { injectTiptapStyles } from './shared/styles'
import { TiptapToolbar } from './components/TiptapToolbar'

export const TiptapRichEditor: React.FC<TiptapRichEditorProps> = ({
  content = '',
  onChange,
  readOnly = false,
  placeholder = 'Start writing...',
  autoFocus = false,
  buttons,
  containerStyle
}) => {
  const isUpdatingFromProps = useRef(false)

  // Inject CSS styles
  useEffect(() => {
    injectTiptapStyles()
  }, [])

  // Memoize the onChange callback to prevent recreating the editor
  const handleContentChange = useCallback(
    (htmlContent: string) => {
      if (onChange && !readOnly && !isUpdatingFromProps.current) {
        onChange(htmlContent)
      }
    },
    [onChange, readOnly]
  )

  const editor = useEditor(
    {
      extensions: createRichExtensions(),
      content,
      editable: !readOnly,
      editorProps: {
        attributes: {
          'data-placeholder': placeholder
        }
      },
      onUpdate: ({ editor }) => {
        const htmlContent = editor.getHTML()
        handleContentChange(htmlContent)
      },
      // Enable content checking for better HTML parsing
      enableContentCheck: true,
      onContentError: ({ editor, error, disableCollaboration }) => {
        console.error('Tiptap content error detected:', error?.message || error)
        // Continue despite content errors
      }
    },
    [handleContentChange, readOnly, placeholder]
  )

  // Update content when prop changes (but avoid loops)
  useEffect(() => {
    if (editor && content !== editor.getHTML()) {
      isUpdatingFromProps.current = true
      editor.commands.setContent(content, false) // false = don't emit update
      // Reset the flag after a short delay to allow for any async operations
      setTimeout(() => {
        isUpdatingFromProps.current = false
      }, 0)
    }
  }, [content, editor])

  // Update readOnly state
  useEffect(() => {
    if (editor) {
      editor.setEditable(!readOnly)
    }
  }, [readOnly, editor])

  // Auto-focus the editor when autoFocus is true and editor is ready
  useEffect(() => {
    if (editor && autoFocus && !readOnly) {
      // Small delay to ensure the editor is fully rendered
      const timer = setTimeout(() => {
        editor.commands.focus('end') // Focus at the end of content
      }, 50)

      return () => clearTimeout(timer)
    }
  }, [editor, autoFocus, readOnly])

  if (!editor) {
    return null
  }

  return (
    <div style={containerStyle}>
      {!readOnly && <TiptapToolbar editor={editor} buttons={buttons} mode="rich" />}
      <EditorContent
        editor={editor}
        style={{
          border: 'none',
          outline: 'none'
        }}
      />
    </div>
  )
}

export default TiptapRichEditor
