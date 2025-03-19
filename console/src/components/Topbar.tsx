import { Layout } from 'antd'
import { LogoutOutlined } from '@ant-design/icons'
import { useAuth } from '../contexts/AuthContext'
import logo from '../assets/logo.png'

const { Header } = Layout

export function Topbar() {
  const { signout } = useAuth()

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
        <img src={logo} alt="Logo" style={{ height: '32px' }} />
      </div>
      <div>
        <LogoutOutlined
          onClick={() => signout()}
          style={{
            fontSize: '18px',
            cursor: 'pointer',
            color: '#595959'
          }}
        />
      </div>
    </Header>
  )
}
