import { useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useNavigate } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'

function Logout() {
  const { logout } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    logout()
    navigate({ to: '/signin' })
  }, [logout, navigate])

  return null
}

export { Logout }
export const Route = createFileRoute('/logout')({
  component: Logout
})
