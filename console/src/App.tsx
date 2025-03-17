import { RouterProvider } from '@tanstack/react-router'
import { ConfigProvider, App as AntApp } from 'antd'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { router } from './router'
import { AuthProvider } from './contexts/AuthContext'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1
    }
  }
})

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AntApp>
        <ConfigProvider
          theme={{
            token: {
              colorPrimary: '#1677ff'
            }
          }}
        >
          <AuthProvider>
            <RouterProvider router={router} />
          </AuthProvider>
        </ConfigProvider>
      </AntApp>
    </QueryClientProvider>
  )
}

export default App
