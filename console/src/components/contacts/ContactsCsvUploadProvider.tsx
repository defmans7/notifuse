import React, { useState } from 'react'
import { Button } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { List } from '../../services/api/types'
import { ContactsCsvUploadDrawer } from './ContactsCsvUploadDrawer'

// Create a context for the singleton
export const CsvUploadContext = React.createContext<{
  openDrawer: (workspaceId: string, lists?: List[]) => void
} | null>(null)

interface ContactsCsvUploadDrawerProviderProps {
  children: React.ReactNode
  workspaceId?: string
}

export const ContactsCsvUploadProvider: React.FC<ContactsCsvUploadDrawerProviderProps> = ({
  children
}) => {
  const [drawerVisible, setDrawerVisible] = useState(false)
  const [contextLists, setContextLists] = useState<List[]>([])
  const [contextWorkspaceId, setContextWorkspaceId] = useState<string>('')

  const openDrawer = (workspaceId: string, lists: List[] = []) => {
    setContextWorkspaceId(workspaceId)
    setContextLists(lists)
    setDrawerVisible(true)
  }

  return (
    <CsvUploadContext.Provider value={{ openDrawer }}>
      {children}
      {drawerVisible && (
        <ContactsCsvUploadDrawer
          workspaceId={contextWorkspaceId}
          lists={contextLists}
          onSuccess={() => {}}
          isVisible={drawerVisible}
          onClose={() => setDrawerVisible(false)}
        />
      )}
    </CsvUploadContext.Provider>
  )
}

export function useContactsCsvUpload() {
  const context = React.useContext(CsvUploadContext)
  if (!context) {
    throw new Error('useContactsCsvUpload must be used within a ContactsCsvUploadDrawerProvider')
  }
  return context
}

export function ContactsCsvUploadButton() {
  const { openDrawer } = useContactsCsvUpload()

  return (
    <Button type="primary" onClick={() => openDrawer('')} icon={<UploadOutlined />}>
      Import Contacts from CSV
    </Button>
  )
}
