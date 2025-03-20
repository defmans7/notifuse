import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { authService } from '../services/api/auth'
import { Workspace } from '../services/api/types'

export interface User {
  id: string
  email: string
}

interface AuthContextType {
  user: User | null
  workspaces: Workspace[]
  isAuthenticated: boolean
  signin: (token: string) => Promise<void>
  signout: () => Promise<void>
  loading: boolean
  refreshWorkspaces: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check for existing session on component mount
    checkAuth()
  }, [])

  const checkAuth = async () => {
    // console.log('checkAuth')
    try {
      // Check if a token exists in localStorage
      const token = localStorage.getItem('auth_token')
      if (!token) {
        setLoading(false)
        return
      }

      // Token exists, fetch current user data
      const { user, workspaces } = await authService.getCurrentUser()
      setUser(user)
      setWorkspaces(workspaces)
      setLoading(false)
    } catch (error) {
      // If there's an error (like an expired token), clear the storage
      localStorage.removeItem('auth_token')
      setUser(null)
      setWorkspaces([])
      setLoading(false)
    }
  }

  const signin = async (token: string) => {
    // console.log('signin')
    try {
      // Store token in localStorage for persistence
      localStorage.setItem('auth_token', token)

      // Fetch current user data using the token
      const { user, workspaces } = await authService.getCurrentUser()
      setUser(user)
      setWorkspaces(workspaces)
    } catch (error) {
      // If there's an error, clear the storage
      localStorage.removeItem('auth_token')
      throw error
    }
  }

  const signout = async () => {
    // Remove token from localStorage
    localStorage.removeItem('auth_token')

    // Clear user data
    setUser(null)
    setWorkspaces([])
  }

  const refreshWorkspaces = async () => {
    const { workspaces } = await authService.getCurrentUser()
    setWorkspaces(workspaces)
  }

  // console.log('user', user)

  return (
    <AuthContext.Provider
      value={{
        user,
        workspaces,
        isAuthenticated: !!user,
        signin,
        signout,
        loading,
        refreshWorkspaces
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
