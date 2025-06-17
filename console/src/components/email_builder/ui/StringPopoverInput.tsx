import React, { useState } from 'react'
import { Input, Popover, Button } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faExternalLinkAlt } from '@fortawesome/free-solid-svg-icons'

interface StringPopoverInputProps {
  value?: string
  onChange: (value: string | undefined) => void
  placeholder?: string
  buttonText?: string
  validateUri?: boolean
}

const StringPopoverInput: React.FC<StringPopoverInputProps> = ({
  value,
  onChange,
  placeholder = 'Enter value',
  buttonText = 'Set value',
  validateUri = false
}) => {
  const [open, setOpen] = useState(false)
  const [inputValue, setInputValue] = useState(value || '')

  const isValidUri = (uri: string): boolean => {
    if (!validateUri || !uri.trim()) return true
    try {
      new URL(uri)
      return true
    } catch {
      return false
    }
  }

  const isCurrentValueValid = isValidUri(inputValue)

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen)
    if (newOpen) {
      setInputValue(value || '')
    }
  }

  const handleSave = () => {
    const trimmedValue = inputValue.trim()
    if (validateUri && trimmedValue && !isValidUri(trimmedValue)) {
      return // Don't save if validation fails
    }
    onChange(trimmedValue || undefined)
    setOpen(false)
  }

  const handleCancel = () => {
    setInputValue(value || '')
    setOpen(false)
  }

  const handleClear = () => {
    onChange(undefined)
    setOpen(false)
  }

  const content = (
    <div className="w-64">
      <Input
        size="small"
        value={inputValue}
        onChange={(e) => setInputValue(e.target.value)}
        placeholder={placeholder}
        onPressEnter={handleSave}
        autoFocus
        allowClear
        onClear={handleClear}
        status={validateUri && inputValue && !isCurrentValueValid ? 'error' : undefined}
      />
      {validateUri && inputValue && !isCurrentValueValid && (
        <div className="text-xs text-red-500 mt-1">Invalid URL format</div>
      )}
      <div className="flex justify-end gap-2 mt-2">
        <Button size="small" onClick={handleCancel}>
          Cancel
        </Button>
        <Button
          size="small"
          type="primary"
          onClick={handleSave}
          disabled={!!(validateUri && inputValue && !isCurrentValueValid)}
        >
          Save
        </Button>
      </div>
    </div>
  )

  if (value) {
    const isValueValid = isValidUri(value)
    const shouldRenderAsLink = validateUri && isValueValid && value.trim()

    return (
      <div className="space-y-2">
        {shouldRenderAsLink ? (
          <div className="flex items-center gap-1">
            <a
              href={value}
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-gray-900 hover:text-blue-500 underline break-all flex-1"
            >
              {value}
            </a>
            <FontAwesomeIcon
              icon={faExternalLinkAlt}
              className="text-xs text-gray-400 opacity-70"
            />
          </div>
        ) : (
          <span
            className={`text-xs block break-all ${
              validateUri && !isValueValid ? 'text-red-500' : 'text-slate-600'
            }`}
          >
            {value}
          </span>
        )}

        <Popover
          content={content}
          title="Edit value"
          trigger="click"
          open={open}
          onOpenChange={handleOpenChange}
          placement="bottom"
        >
          <Button type="primary" size="small" ghost>
            Edit value
          </Button>
        </Popover>
      </div>
    )
  }

  return (
    <Popover
      content={content}
      title="Set value"
      trigger="click"
      open={open}
      onOpenChange={handleOpenChange}
      placement="bottom"
    >
      <Button size="small" type="primary" ghost className="text-xs">
        {buttonText}
      </Button>
    </Popover>
  )
}

export default StringPopoverInput
