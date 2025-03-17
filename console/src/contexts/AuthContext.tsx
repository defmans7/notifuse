import { createContext, useContext, useState, ReactNode, useEffect } from 'react'
import config from '../config'

interface User {
  email: string
  timezone: string
}

interface Workspace {
  id: string
  settings: {
    name: string
    url: string
    logo_url: string | null
    timezone: string
  }
}

interface AuthContextType {
  isAuthenticated: boolean
  user: User | null
  workspaces: Workspace[]
  login: (token: string, user: User) => Promise<void>
  logout: () => void
  refreshing: boolean
  refreshWorkspaces: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | null>(null)

const fetchWorkspaces = async (): Promise<Workspace[]> => {
  const authToken = localStorage.getItem('auth_token')
  if (!authToken) {
    throw new Error('No authentication token')
  }

  const response = await fetch(`${config.API_ENDPOINT}/api/workspaces.list`, {
    headers: {
      Authorization: `Bearer ${authToken}`
    }
  })

  if (response.status === 401) {
    // Force logout on unauthorized
    throw new Error('Unauthorized')
  }

  if (!response.ok) {
    throw new Error('Failed to fetch workspaces')
  }

  return response.json()
}

function LoadingScreen() {
  return (
    <div className="flex h-screen w-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
    </div>
  )
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => {
    const authToken = localStorage.getItem('auth_token')
    return !!authToken
  })
  const [user, setUser] = useState<User | null>(() => {
    const userData = localStorage.getItem('user')
    return userData ? JSON.parse(userData) : null
  })
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [refreshing, setRefreshing] = useState(false)
  const [initialized, setInitialized] = useState(false)

  const refreshWorkspaces = async () => {
    setRefreshing(true)
    try {
      const data = await fetchWorkspaces()
      console.log('refreshWorkspaces', data)
      setWorkspaces(data)
    } catch (error) {
      console.error('Failed to fetch workspaces:', error)
      setWorkspaces([])
      if (error instanceof Error && error.message === 'Unauthorized') {
        logout()
      }
    } finally {
      setRefreshing(false)
    }
  }

  const login = async (token: string, userData: User) => {
    console.log('login', token, userData)
    localStorage.setItem('auth_token', token)
    localStorage.setItem('user', JSON.stringify(userData))
    await refreshWorkspaces()
    setIsAuthenticated(true)
    setUser(userData)
    setInitialized(true)
  }

  const logout = () => {
    console.log('logout')
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user')
    setIsAuthenticated(false)
    setUser(null)
    setWorkspaces([])
    setInitialized(true)
  }

  // Initialize auth state from localStorage
  useEffect(() => {
    const authToken = localStorage.getItem('auth_token')
    const userData = localStorage.getItem('user')

    if (authToken && userData) {
      try {
        const user = JSON.parse(userData)
        // We don't await this since login already handles setting initialized
        login(authToken, user)
      } catch (error) {
        console.error('Failed to parse stored user data:', error)
        logout()
      }
    } else {
      setInitialized(true)
    }
  }, []) // Empty deps since we only want this to run once on mount

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        user,
        workspaces,
        login,
        logout,
        refreshing,
        refreshWorkspaces
      }}
    >
      {!initialized ? <LoadingScreen /> : children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
