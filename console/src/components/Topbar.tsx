import { Layout } from 'antd'
import { useNavigate } from '@tanstack/react-router'
import logo from '../assets/logo.png'

const { Header } = Layout

export function Topbar() {
  const navigate = useNavigate()

  return (
    <Header
      style={{
        padding: '0 24px',
        background: '#fff',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)'
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center' }}>
        <img
          src={logo}
          alt="Logo"
          style={{ height: '32px', cursor: 'pointer' }}
          onClick={() => navigate({ to: '/' })}
        />
      </div>
    </Header>
  )
}
