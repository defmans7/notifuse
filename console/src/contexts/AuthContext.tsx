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
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check for existing session
    checkAuth()
  }, [])

  const checkAuth = async () => {
    try {
      // TODO: Implement actual session check
      setLoading(false)
    } catch (error) {
      setLoading(false)
    }
  }

  const signin = async (token: string) => {
    try {
      // TODO: Implement actual signin
      const currentUserResponse = await authService.getCurrentUser()
      const { user, workspaces } = currentUserResponse
      // Mock user for now
      setUser(user)
      // Mock workspaces fetch
      setWorkspaces(workspaces)
    } catch (error) {
      throw error
    }
  }

  const signout = async () => {
    // TODO: Implement actual signout
    setUser(null)
    setWorkspaces([])
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        workspaces,
        isAuthenticated: !!user,
        signin,
        signout,
        loading
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
