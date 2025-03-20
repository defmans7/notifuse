import { Layout } from 'antd'
import { ReactNode } from 'react'
import { Topbar } from '../components/Topbar'

const { Content } = Layout

interface MainLayoutProps {
  children: ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Topbar />
      <Content style={{ padding: '24px' }}>{children}</Content>
    </Layout>
  )
}
