import { useEffect, useRef } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useNavigate } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'

function Logout() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const hasLoggedOut = useRef(false)

  useEffect(() => {
    if (!hasLoggedOut.current) {
      hasLoggedOut.current = true
      logout()
      navigate({ to: '/signin' })
    }
  }, [logout, navigate])

  return null
}

export { Logout }
export const Route = createFileRoute('/logout')({
  component: Logout
})
