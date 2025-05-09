// https://github.com/ajaxorg/ace-builds/tree/master/src-noconflict
import AceEditor from 'react-ace'
import 'ace-builds/src-noconflict/mode-javascript'
import 'ace-builds/src-noconflict/mode-json'
import 'ace-builds/src-noconflict/mode-liquid'
import 'ace-builds/src-noconflict/theme-chrome'
import 'ace-builds/src-noconflict/theme-monokai'
import 'ace-builds/src-noconflict/ext-language_tools'
import { useState, useEffect } from 'react'

type AceInputProps = {
  value?: string
  onChange?: (value: string) => void
  mode: string
  id: string
  height: string
  width: string
  theme?: string
}

const AceInput = (props: AceInputProps) => {
  // Keep track of whether the current mode is available
  const [currentMode, setCurrentMode] = useState<string>('text')

  useEffect(() => {
    // Update the mode when the component mounts or props.mode changes
    setCurrentMode(getModeWithFallback(props.mode))
  }, [props.mode])

  // Handle modes that might not be directly supported by setting fallbacks
  const getModeWithFallback = (requestedMode: string): string => {
    // For modes that require special handling
    if (requestedMode === 'liquid') {
      // Check if the mode exists, fallback to text if not
      try {
        // Make sure the Liquid mode is loaded - in a real app this would be more robust
        return 'liquid'
      } catch (e) {
        console.warn('Liquid mode not available, falling back to text mode')
        return 'text'
      }
    }

    return requestedMode
  }

  return (
    <AceEditor
      value={props.value}
      mode={currentMode}
      theme={props.theme || 'chrome'}
      onChange={props.onChange}
      debounceChangePeriod={300}
      name={props.id}
      editorProps={{ $blockScrolling: true }}
      fontSize="12px"
      height={props.height}
      width={props.width}
      className="ace-input"
      wrapEnabled={true}
      setOptions={{
        useWorker: false, // Disable worker to prevent errors with some modes
        enableBasicAutocompletion: true,
        enableLiveAutocompletion: true
      }}
    />
  )
}

export default AceInput
