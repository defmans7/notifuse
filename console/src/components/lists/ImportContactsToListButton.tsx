import React from 'react'
import { Button, Tooltip } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { useContactsCsvUpload } from '../contacts/ContactsCsvUploadProvider'
import { List } from '../../services/api/types'
import { useQueryClient } from '@tanstack/react-query'

interface ImportContactsToListButtonProps {
  list: List
  workspaceId: string
  lists?: List[]
  size?: 'large' | 'middle' | 'small'
  type?: 'default' | 'primary' | 'dashed' | 'link' | 'text'
  className?: string
  style?: React.CSSProperties
}

export function ImportContactsToListButton({
  list,
  workspaceId,
  lists = [],
  size = 'small',
  type = 'text',
  className,
  style
}: ImportContactsToListButtonProps) {
  const { openDrawerWithSelectedList } = useContactsCsvUpload()
  const queryClient = useQueryClient()

  const handleClick = () => {
    // Pass true for refreshOnClose to refresh contacts data
    openDrawerWithSelectedList(workspaceId, lists, list.id, true)

    // Listen for the drawer to close with a successful import
    const handleSuccess = () => {
      // Also invalidate lists query to refresh list counts
      queryClient.invalidateQueries({ queryKey: ['lists', workspaceId] })
    }

    // Set up a one-time event listener for the import success
    document.addEventListener('contactsImported', handleSuccess, { once: true })
  }

  return (
    <Button type={type} size={size} onClick={handleClick} className={className} style={style}>
      <Tooltip title="Import Contacts to List">
        <UploadOutlined />
      </Tooltip>
    </Button>
  )
}
